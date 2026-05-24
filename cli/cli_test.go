package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestListTypes(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Run(context.Background(), []string{"list-types"}, &out, &errOut)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "cmb_debit") {
		t.Fatalf("missing adapter list: %s", out.String())
	}
	if !strings.Contains(out.String(), "abc_debit") {
		t.Fatalf("missing abc_debit in adapter list: %s", out.String())
	}
}

func TestConfigInitWritesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	var out, errOut bytes.Buffer
	code := Run(context.Background(), []string{"config", "init", "-output", path}, &out, &errOut)
	if code != 0 {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	for _, want := range []string{"mineru:", "base_url:", "lang_list:", "timeout: 10m"} {
		if !strings.Contains(text, want) {
			t.Fatalf("config missing %q: %s", want, text)
		}
	}
	for _, forbidden := range []string{"formula_enable", "table_enable", "return_content_list", "return_md"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("config should not expose %q: %s", forbidden, text)
		}
	}
}

func TestLoadConfigYAMLDefaultsAndValidation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(`mineru:
  base_url: "http://192.168.1.20:8000"
`), 0o644); err != nil {
		t.Fatal(err)
	}
	config, err := LoadConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	minerU, err := config.MinerUHTTPConfig()
	if err != nil {
		t.Fatal(err)
	}
	if minerU.BaseURL != "http://192.168.1.20:8000" || minerU.ParseMethod != "auto" || minerU.Backend != "hybrid-auto-engine" {
		t.Fatalf("unexpected config: %#v", minerU)
	}
}

func TestConvertRequiresPDFArgument(t *testing.T) {
	var out, errOut bytes.Buffer
	code := Run(context.Background(), []string{"convert", "-type", "cmb_debit"}, &out, &errOut)
	if code == 0 || !strings.Contains(errOut.String(), "expected at least one PDF file or directory") {
		t.Fatalf("code=%d stderr=%s", code, errOut.String())
	}
}

func TestNormalizeConvertFlagsAfterPDF(t *testing.T) {
	args := normalizeFlagArgs([]string{
		"/tmp/statement.pdf",
		"--type", "cmb_debit",
		"--config=config.yaml",
		"--output", "output",
	}, map[string]bool{
		"-type": true, "--type": true,
		"-config": true, "--config": true,
		"-output": true, "--output": true,
	})
	want := []string{"--type", "cmb_debit", "--config=config.yaml", "--output", "output", "/tmp/statement.pdf"}
	if strings.Join(args, "\n") != strings.Join(want, "\n") {
		t.Fatalf("normalized args = %#v, want %#v", args, want)
	}
}

func TestExpandInputPathsDirectoryNaturalOrder(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"10.pdf", "2.pdf", "1.pdf", "note.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	standalone := filepath.Join(t.TempDir(), "standalone.pdf")
	if err := os.WriteFile(standalone, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	paths, err := expandInputPaths([]string{standalone, dir})
	if err != nil {
		t.Fatal(err)
	}
	got := []string{}
	for _, path := range paths {
		got = append(got, filepath.Base(path))
	}
	want := []string{"standalone.pdf", "1.pdf", "2.pdf", "10.pdf"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
}

func TestNaturalLess(t *testing.T) {
	if !naturalLess("2.pdf", "10.pdf") {
		t.Fatal("expected natural ordering")
	}
	if !naturalLess("a1.pdf", "a2.pdf") {
		t.Fatal("expected lexical prefix ordering")
	}
}
