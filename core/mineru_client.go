package core

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type MinerUHTTPConfig struct {
	BaseURL     string
	LangList    []string
	Backend     string
	ParseMethod string
	Timeout     time.Duration
	MaxRetries  int
	Headers     map[string]string
}

type MinerUHTTPClient struct {
	config MinerUHTTPConfig
	client *http.Client
}

func NewMinerUHTTPClient(config MinerUHTTPConfig) (*MinerUHTTPClient, error) {
	if err := validateMinerUHTTPConfig(config); err != nil {
		return nil, err
	}
	return &MinerUHTTPClient{
		config: config,
		client: &http.Client{Timeout: config.Timeout},
	}, nil
}

func validateMinerUHTTPConfig(config MinerUHTTPConfig) error {
	baseURL := strings.TrimSpace(config.BaseURL)
	if baseURL == "" {
		return fmt.Errorf("mineru.base_url is required")
	}
	if strings.Contains(baseURL, "<") || strings.Contains(baseURL, ">") {
		return fmt.Errorf("mineru.base_url still contains a placeholder: %q", baseURL)
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return fmt.Errorf("mineru.base_url must be an absolute http(s) URL")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("mineru.base_url must use http or https")
	}
	return nil
}

func (c *MinerUHTTPClient) Ping(ctx context.Context) error {
	endpoint := joinURL(c.config.BaseURL, "/health")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	for key, value := range c.config.Headers {
		req.Header.Set(key, value)
	}
	client := c.client
	if client.Timeout == 0 {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	body, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("MinerU returned %s: %s", resp.Status, strings.TrimSpace(string(body)))
}

func (c *MinerUHTTPClient) Parse(ctx context.Context, file InputFile) (MinerUParseResult, error) {
	rawRequest := rawMinerURequest(c.config, file)
	file, err := retryableInputFile(file)
	if err != nil {
		return MinerUParseResult{RawRequest: rawRequest}, err
	}
	var last MinerUParseResult
	attempts := c.config.MaxRetries + 1
	if attempts < 1 {
		attempts = 1
	}
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return last, ctx.Err()
			case <-time.After(backoffDelay(attempt)):
			}
		}
		result, status, err := c.parseOnce(ctx, file, rawRequest)
		last = result
		lastErr = err
		if err == nil {
			return result, nil
		}
		if status > 0 && status < 500 && status != http.StatusTooManyRequests {
			return result, err
		}
	}
	return last, lastErr
}

func retryableInputFile(file InputFile) (InputFile, error) {
	if file.Reader == nil {
		return file, nil
	}
	if _, ok := file.Reader.(io.Seeker); ok {
		return file, nil
	}
	data, err := io.ReadAll(file.Reader)
	if err != nil {
		return file, err
	}
	file.Reader = bytes.NewReader(data)
	return file, nil
}

func (c *MinerUHTTPClient) parseOnce(ctx context.Context, file InputFile, rawRequest string) (MinerUParseResult, int, error) {
	body, contentType, err := c.multipartBody(file)
	if err != nil {
		return MinerUParseResult{RawRequest: rawRequest}, 0, err
	}
	endpoint := joinURL(c.config.BaseURL, "/file_parse")
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return MinerUParseResult{RawRequest: rawRequest}, 0, err
	}
	req.Header.Set("Content-Type", contentType)
	for key, value := range c.config.Headers {
		req.Header.Set(key, value)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return MinerUParseResult{RawRequest: rawRequest}, 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return MinerUParseResult{RawRequest: rawRequest, RawResponse: string(data)}, resp.StatusCode, err
	}
	result := MinerUParseResult{RawRequest: rawRequest, RawResponse: string(data)}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return result, resp.StatusCode, fmt.Errorf("MinerU returned %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}
	contentList, err := DecodeMinerUContentList(data)
	if err != nil {
		return result, resp.StatusCode, err
	}
	result.ContentList = contentList
	return result, resp.StatusCode, nil
}

func (c *MinerUHTTPClient) multipartBody(file InputFile) ([]byte, string, error) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	if err := addMultipartFile(writer, file); err != nil {
		_ = writer.Close()
		return nil, "", err
	}
	for _, lang := range c.config.LangList {
		if err := writer.WriteField("lang_list", lang); err != nil {
			_ = writer.Close()
			return nil, "", err
		}
	}
	fields := map[string]string{
		"backend":             c.config.Backend,
		"parse_method":        c.config.ParseMethod,
		"return_md":           "false",
		"formula_enable":      "false",
		"table_enable":        "true",
		"return_content_list": "true",
	}
	for key, value := range fields {
		if value == "" {
			continue
		}
		if err := writer.WriteField(key, value); err != nil {
			_ = writer.Close()
			return nil, "", err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), writer.FormDataContentType(), nil
}

func addMultipartFile(writer *multipart.Writer, file InputFile) error {
	part, err := writer.CreateFormFile("files", file.Name())
	if err != nil {
		return err
	}
	if file.Reader != nil {
		if seeker, ok := file.Reader.(io.Seeker); ok {
			if _, err := seeker.Seek(0, io.SeekStart); err != nil {
				return err
			}
		}
		_, err = io.Copy(part, file.Reader)
		return err
	}
	f, err := os.Open(file.Path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(part, f)
	return err
}

func DecodeMinerUContentList(data []byte) ([]MinerUContent, error) {
	var payload any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("decode MinerU response JSON: %w", err)
	}
	rawLists := collectContentLists(payload)
	if len(rawLists) == 0 {
		return nil, fmt.Errorf("MinerU response did not include content_list")
	}
	var combined []MinerUContent
	for _, raw := range rawLists {
		items, err := decodeContentListValue(raw)
		if err != nil {
			return nil, err
		}
		combined = append(combined, items...)
	}
	return combined, nil
}

func collectContentLists(value any) []any {
	switch typed := value.(type) {
	case map[string]any:
		var lists []any
		if raw, ok := typed["content_list"]; ok {
			lists = append(lists, raw)
		}
		for key, raw := range typed {
			if key == "content_list" {
				continue
			}
			lists = append(lists, collectContentLists(raw)...)
		}
		return lists
	case []any:
		var lists []any
		for _, item := range typed {
			lists = append(lists, collectContentLists(item)...)
		}
		return lists
	}
	return nil
}

func decodeContentListValue(value any) ([]MinerUContent, error) {
	if text, ok := value.(string); ok {
		var decoded any
		if err := json.Unmarshal([]byte(text), &decoded); err != nil {
			return nil, fmt.Errorf("decode string content_list: %w", err)
		}
		return decodeContentListValue(decoded)
	}
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var rawItems []map[string]any
	if err := json.Unmarshal(data, &rawItems); err != nil {
		return nil, fmt.Errorf("decode content_list: %w", err)
	}
	items := make([]MinerUContent, 0, len(rawItems))
	for _, raw := range rawItems {
		item := MinerUContent{Raw: raw}
		if value, _ := raw["type"].(string); value != "" {
			item.Type = value
		}
		if value, _ := raw["text"].(string); value != "" {
			item.Text = value
		}
		if value, _ := raw["table_body"].(string); value != "" {
			item.TableBody = value
		}
		if page, ok := rawPageIndex(raw["page_idx"]); ok {
			item.PageIndex = &page
		} else if page, ok := rawPageIndex(raw["page"]); ok {
			page--
			item.PageIndex = &page
		}
		items = append(items, item)
	}
	return items, nil
}

func rawPageIndex(value any) (int, bool) {
	switch typed := value.(type) {
	case float64:
		return int(typed), true
	case int:
		return typed, true
	case string:
		parsed, err := strconv.Atoi(strings.TrimSpace(typed))
		return parsed, err == nil
	default:
		return 0, false
	}
}

func rawMinerURequest(config MinerUHTTPConfig, file InputFile) string {
	payload := map[string]any{
		"method": "POST",
		"url":    joinURL(config.BaseURL, "/file_parse"),
		"form": map[string]any{
			"files":               []string{file.Name()},
			"lang_list":           config.LangList,
			"backend":             config.Backend,
			"parse_method":        config.ParseMethod,
			"return_md":           false,
			"formula_enable":      false,
			"table_enable":        true,
			"return_content_list": true,
		},
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return ""
	}
	return string(data)
}

func joinURL(base, suffix string) string {
	base = strings.TrimRight(base, "/")
	if base == "" {
		return suffix
	}
	return base + "/" + strings.TrimLeft(suffix, "/")
}

func backoffDelay(attempt int) time.Duration {
	base := 500 * time.Millisecond
	max := 8 * time.Second
	delay := base << (attempt - 1)
	if delay > max {
		return max
	}
	return delay
}
