package core

import (
	"context"
	"encoding/json"
	"io"
	"path/filepath"

	"github.com/deb-sig/bill-file-converter/core/adapters"
)

type Input struct {
	Files []InputFile
}

type InputFile struct {
	Path     string
	Reader   io.Reader
	FileName string
	MIMEType string
}

func (f InputFile) Name() string {
	if f.FileName != "" {
		return f.FileName
	}
	if f.Path != "" {
		return filepath.Base(f.Path)
	}
	return "input.pdf"
}

type Options struct {
	MinerU          MinerUClient
	AdapterKey      string
	OutputDir       string
	AdapterRegistry AdapterRegistry
	SkipCSV         bool
	LogWriter       io.Writer
}

type AdapterRegistry interface {
	MustGet(key string) (adapters.Adapter, error)
}

type Document struct {
	Metadata map[string]string `json:"metadata,omitempty"`
	Tables   []Table           `json:"tables"`
}

type SourceInfo struct {
	Files []SourceFileInfo `json:"files,omitempty"`
}

type SourceFileInfo struct {
	Path     string `json:"path,omitempty"`
	FileName string `json:"file_name,omitempty"`
	MIMEType string `json:"mime_type,omitempty"`
}

type Table struct {
	Name        string      `json:"name,omitempty"`
	Headers     []string    `json:"headers"`
	Rows        [][]*string `json:"rows"`
	SourcePages []int       `json:"source_pages,omitempty"`
	Warnings    []string    `json:"warnings,omitempty"`
}

type ValidationReport struct {
	Errors   []string `json:"errors,omitempty"`
	Warnings []string `json:"warnings,omitempty"`
}

func (r ValidationReport) HasErrors() bool {
	return len(r.Errors) > 0
}

type Artifacts struct {
	JSONPath              string `json:"json_path,omitempty"`
	CSVPath               string `json:"csv_path,omitempty"`
	LoggerDir             string `json:"logger_dir,omitempty"`
	ProcessLogPath        string `json:"process_log_path,omitempty"`
	ContentListPath       string `json:"content_list_path,omitempty"`
	MinerURawRequestPath  string `json:"mineru_raw_request_path,omitempty"`
	MinerURawResponsePath string `json:"mineru_raw_response_path,omitempty"`
	FailurePath           string `json:"failure_path,omitempty"`
}

type Result struct {
	TaskID           string            `json:"task_id"`
	AdapterKey       string            `json:"adapter_key"`
	AdapterName      string            `json:"adapter_name"`
	Source           SourceInfo        `json:"source"`
	GeneratedAt      string            `json:"generated_at"`
	Metadata         map[string]string `json:"metadata,omitempty"`
	Tables           []Table           `json:"tables"`
	ValidationReport ValidationReport  `json:"validation_report"`
	Artifacts        Artifacts         `json:"artifacts,omitempty"`
}

type MinerUContent struct {
	Type      string         `json:"type,omitempty"`
	Text      string         `json:"text,omitempty"`
	TableBody string         `json:"table_body,omitempty"`
	PageIndex *int           `json:"page_idx,omitempty"`
	Raw       map[string]any `json:"-"`
}

func (c MinerUContent) MarshalJSON() ([]byte, error) {
	payload := map[string]any{}
	for key, value := range c.Raw {
		payload[key] = value
	}
	if c.Type != "" {
		payload["type"] = c.Type
	}
	if c.Text != "" {
		payload["text"] = c.Text
	}
	if c.TableBody != "" {
		payload["table_body"] = c.TableBody
	}
	if c.PageIndex != nil {
		payload["page_idx"] = *c.PageIndex
	}
	return json.Marshal(payload)
}

type MinerUParseResult struct {
	ContentList []MinerUContent
	RawRequest  string
	RawResponse string
}

type MinerUClient interface {
	Parse(ctx context.Context, file InputFile) (MinerUParseResult, error)
}
