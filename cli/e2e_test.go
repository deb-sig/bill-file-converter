package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

type e2eCase struct {
	Name     string     `json:"name"`
	Type     string     `json:"type"`
	Inputs   []e2eInput `json:"inputs"`
	Expected string     `json:"expected"`
}

type e2eInput struct {
	Filename string `json:"filename"`
	Response string `json:"response"`
}

func TestConvertCLIWithFakeMinerU(t *testing.T) {
	root := repoRoot(t)
	fixtureDir := filepath.Join(root, "testdata", "e2e")
	cases := loadE2ECases(t, fixtureDir)

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			responses := map[string][]byte{}
			for _, input := range tc.Inputs {
				responses[input.Filename] = readFile(t, filepath.Join(fixtureDir, input.Response))
			}

			server := httptest.NewServer(fakeMinerUHandler(t, responses))
			defer server.Close()

			tempDir := t.TempDir()
			configPath := filepath.Join(tempDir, "config.yaml")
			config := fmt.Sprintf(`mineru:
  base_url: %q
  lang_list: ["ch"]
  backend: "hybrid-auto-engine"
  parse_method: "auto"
  timeout: "5s"
  max_retries: 0
  headers: {}
`, server.URL)
			if err := os.WriteFile(configPath, []byte(config), 0o600); err != nil {
				t.Fatal(err)
			}

			inputDir := filepath.Join(tempDir, "input")
			if err := os.MkdirAll(inputDir, 0o755); err != nil {
				t.Fatal(err)
			}
			args := []string{"run", "./cmd/bill-file-converter", "convert"}
			for _, input := range tc.Inputs {
				path := filepath.Join(inputDir, input.Filename)
				if err := os.WriteFile(path, []byte("%PDF-1.7\n"), 0o600); err != nil {
					t.Fatal(err)
				}
				args = append(args, path)
			}
			outDir := filepath.Join(tempDir, "output")
			args = append(args, "--type", tc.Type, "--config", configPath, "--output", outDir)

			cmd := exec.Command(goTool(t), args...)
			cmd.Dir = root
			cmd.Env = append(os.Environ(), "BFC_E2E_SKIP_REMOVE_IMAGES=1")
			output, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("CLI failed: %v\n%s", err, output)
			}

			actual := findResultCSV(t, outDir)
			got := readFile(t, actual)
			want := readFile(t, filepath.Join(fixtureDir, tc.Expected))
			if !bytes.Equal(got, want) {
				t.Fatalf("unexpected CSV for %s\nwant:\n%s\ngot:\n%s", tc.Name, want, got)
			}
		})
	}
}

func TestE2EFixturesDoNotContainSensitiveValues(t *testing.T) {
	root := repoRoot(t)
	fixtureDir := filepath.Join(root, "testdata", "e2e")
	rules := loadSensitiveRules(t, filepath.Join(root, ".private", "e2e-sensitive-rules.json"))

	patterns := []struct {
		name    string
		re      *regexp.Regexp
		capture int
	}{
		{name: "email", re: regexp.MustCompile(`[A-Za-z0-9._%+\-]+@[A-Za-z0-9.\-]+\.[A-Za-z]{2,}`), capture: 0},
		{name: "mobile", re: regexp.MustCompile(`(?:^|[^0-9])(1[3-9][0-9]{9})(?:[^0-9]|$)`), capture: 1},
		{name: "id_card", re: regexp.MustCompile(`(?i)([1-9][0-9]{5}(?:18|19|20)[0-9]{2}(?:0[1-9]|1[0-2])(?:0[1-9]|[12][0-9]|3[01])[0-9]{3}[0-9X])`), capture: 1},
		{name: "long_digits", re: regexp.MustCompile(`[0-9]{12,}`), capture: 0},
	}

	err := filepath.WalkDir(fixtureDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if ext := filepath.Ext(path); ext != ".json" && ext != ".csv" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(data)
		for _, blocked := range rules.Block {
			if blocked != "" && strings.Contains(text, blocked) {
				t.Errorf("%s contains blocked fixture value %q", path, blocked)
			}
		}
		for _, pattern := range patterns {
			for _, match := range pattern.re.FindAllStringSubmatch(text, -1) {
				value := match[pattern.capture]
				if !rules.allowed[value] {
					t.Errorf("%s contains %s-like value %q not listed in allow rules", path, pattern.name, value)
				}
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func fakeMinerUHandler(t *testing.T, responses map[string][]byte) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			w.WriteHeader(http.StatusOK)
			return
		case "/file_parse":
		default:
			http.NotFound(w, r)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseMultipartForm(16 << 20); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		wantFields := map[string]string{
			"return_content_list": "true",
			"table_enable":        "true",
			"return_md":           "false",
			"formula_enable":      "false",
		}
		for key, want := range wantFields {
			if got := r.FormValue(key); got != want {
				http.Error(w, fmt.Sprintf("%s = %q, want %q", key, got, want), http.StatusBadRequest)
				return
			}
		}
		files := r.MultipartForm.File["files"]
		if len(files) != 1 {
			http.Error(w, fmt.Sprintf("got %d files, want 1", len(files)), http.StatusBadRequest)
			return
		}
		file, err := files[0].Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = io.Copy(io.Discard, file)
		_ = file.Close()

		response, ok := responses[files[0].Filename]
		if !ok {
			http.Error(w, fmt.Sprintf("unknown fixture file %q", files[0].Filename), http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(response)
	})
}

func loadE2ECases(t *testing.T, fixtureDir string) []e2eCase {
	t.Helper()
	var cases []e2eCase
	if err := json.Unmarshal(readFile(t, filepath.Join(fixtureDir, "cases.json")), &cases); err != nil {
		t.Fatal(err)
	}
	if len(cases) == 0 {
		t.Fatal("no e2e cases configured")
	}
	return cases
}

type sensitiveRules struct {
	Allow   []string `json:"allow"`
	Block   []string `json:"block"`
	allowed map[string]bool
}

func loadSensitiveRules(t *testing.T, path string) sensitiveRules {
	t.Helper()
	// Optional local-only rules file, ignored by git/AI tools:
	//
	//   {
	//     "allow": ["known-sanitized-value-that-matches-a-generic-pattern"],
	//     "block": ["real-name-or-account-that-must-never-appear"]
	//   }
	//
	// Block entries always fail. Allow entries only exempt generic pattern hits.
	rules := sensitiveRules{allowed: map[string]bool{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return rules
		}
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &rules); err != nil {
		t.Fatal(err)
	}
	rules.allowed = map[string]bool{}
	for _, value := range rules.Allow {
		rules.allowed[value] = true
	}
	return rules
}

func findResultCSV(t *testing.T, outDir string) string {
	t.Helper()
	var matches []string
	err := filepath.WalkDir(outDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(filepath.ToSlash(path), "/result/result.csv") {
			matches = append(matches, path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 1 {
		t.Fatalf("found %d result.csv files under %s: %v", len(matches), outDir, matches)
	}
	return matches[0]
}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func goTool(t *testing.T) string {
	t.Helper()
	path := filepath.Join(runtime.GOROOT(), "bin", "go")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return "go"
}
