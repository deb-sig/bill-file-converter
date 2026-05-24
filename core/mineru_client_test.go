package core

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMinerUHTTPClientPostsMultipartToFileParse(t *testing.T) {
	dir := t.TempDir()
	pdf := filepath.Join(dir, "1.pdf")
	if err := os.WriteFile(pdf, []byte("%PDF-1.7"), 0o644); err != nil {
		t.Fatal(err)
	}
	var seenPath string
	var seenFiles int
	var seenReturnContentList string
	var seenReturnMD string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenPath = r.URL.Path
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		seenFiles = len(r.MultipartForm.File["files"])
		seenReturnContentList = r.MultipartForm.Value["return_content_list"][0]
		seenReturnMD = r.MultipartForm.Value["return_md"][0]
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content_list": []map[string]any{{"type": "text", "text": "ok"}},
		})
	}))
	defer server.Close()
	client, err := NewMinerUHTTPClient(MinerUHTTPConfig{
		BaseURL:     server.URL,
		LangList:    []string{"ch"},
		Backend:     "hybrid-auto-engine",
		ParseMethod: "auto",
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Parse(context.Background(), InputFile{Path: pdf, FileName: "1.pdf"})
	if err != nil {
		t.Fatal(err)
	}
	if seenPath != "/file_parse" {
		t.Fatalf("path = %q", seenPath)
	}
	if seenFiles != 1 {
		t.Fatalf("files = %d", seenFiles)
	}
	if seenReturnContentList != "true" {
		t.Fatalf("return_content_list = %q", seenReturnContentList)
	}
	if seenReturnMD != "false" {
		t.Fatalf("return_md = %q", seenReturnMD)
	}
	if len(result.ContentList) != 1 || result.ContentList[0].Text != "ok" {
		t.Fatalf("unexpected content list: %#v", result.ContentList)
	}
}

func TestMinerUHTTPClientRetriesReaderBackedFileWithFullBody(t *testing.T) {
	const pdfBody = "%PDF-1.7 reader backed input"
	var uploads []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(32 << 20); err != nil {
			t.Fatalf("parse multipart: %v", err)
		}
		files := r.MultipartForm.File["files"]
		if len(files) != 1 {
			t.Fatalf("files = %d", len(files))
		}
		file, err := files[0].Open()
		if err != nil {
			t.Fatalf("open upload: %v", err)
		}
		data, err := io.ReadAll(file)
		_ = file.Close()
		if err != nil {
			t.Fatalf("read upload: %v", err)
		}
		uploads = append(uploads, string(data))
		if len(uploads) == 1 {
			http.Error(w, "try again", http.StatusInternalServerError)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"content_list": []map[string]any{{"type": "text", "text": "ok"}},
		})
	}))
	defer server.Close()
	client, err := NewMinerUHTTPClient(MinerUHTTPConfig{
		BaseURL:    server.URL,
		MaxRetries: 1,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := client.Parse(context.Background(), InputFile{
		Reader:   io.NopCloser(strings.NewReader(pdfBody)),
		FileName: "reader.pdf",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.ContentList) != 1 || result.ContentList[0].Text != "ok" {
		t.Fatalf("unexpected content list: %#v", result.ContentList)
	}
	if len(uploads) != 2 {
		t.Fatalf("uploads = %d", len(uploads))
	}
	for i, upload := range uploads {
		if upload != pdfBody {
			t.Fatalf("upload %d = %q", i, upload)
		}
	}
}

func TestDecodeMinerUContentListShapes(t *testing.T) {
	cases := []string{
		`{"content_list":[{"type":"text","text":"a"}]}`,
		`{"data":{"content_list":[{"type":"text","text":"a"}]}}`,
		`{"data":{"content_list":"[{\"type\":\"text\",\"text\":\"a\"}]"}}`,
	}
	for _, tc := range cases {
		items, err := DecodeMinerUContentList([]byte(tc))
		if err != nil {
			t.Fatalf("%s: %v", tc, err)
		}
		if len(items) != 1 || items[0].Text != "a" {
			t.Fatalf("%s: %#v", tc, items)
		}
	}
}

func TestDecodeMinerUContentListCombinesMultipleResults(t *testing.T) {
	items, err := DecodeMinerUContentList([]byte(`{
		"results": [
			{"content_list": [{"type":"text","text":"a"}]},
			{"content_list": [{"type":"text","text":"b"}]}
		]
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Text != "a" || items[1].Text != "b" {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestDecodeMinerUContentListFromNamedFileResults(t *testing.T) {
	items, err := DecodeMinerUContentList([]byte(`{
		"status": "completed",
		"results": {
			"test": {
				"content_list": "[{\"type\":\"text\",\"text\":\"a\"}]"
			}
		}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].Text != "a" {
		t.Fatalf("unexpected items: %#v", items)
	}
}

func TestMinerUHTTPConfigValidation(t *testing.T) {
	for _, baseURL := range []string{"", "http://127.0.0.1:<port>", "ftp://host", "127.0.0.1:8000"} {
		if _, err := NewMinerUHTTPClient(MinerUHTTPConfig{BaseURL: baseURL}); err == nil {
			t.Fatalf("expected invalid base_url %q", baseURL)
		}
	}
	for _, baseURL := range []string{"http://127.0.0.1:12345", "http://192.168.1.20:8000", "https://mineru.example.internal"} {
		if _, err := NewMinerUHTTPClient(MinerUHTTPConfig{BaseURL: baseURL}); err != nil {
			t.Fatalf("expected valid base_url %q: %v", baseURL, err)
		}
	}
}
