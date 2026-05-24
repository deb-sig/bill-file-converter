package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// NewTaskLogger creates the business logger used by one Convert run.
func NewTaskLogger(outputDir string, taskID string, adapterKey string, terminal io.Writer) (*Logger, error) {
	return newBase(outputDir, taskID, adapterKey, terminal)
}

type failureArtifact struct {
	Time       string `json:"time"`
	TaskID     string `json:"task_id"`
	Source     any    `json:"source"`
	AdapterKey string `json:"adapter_key,omitempty"`
	Adapter    string `json:"adapter,omitempty"`
	Error      string `json:"error"`
}

func (l *Logger) Verbosef(format string, args ...any) {
	l.logf(Verbose, format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.logf(Info, format, args...)
}

func (l *Logger) Warningf(format string, args ...any) {
	l.logf(Warning, format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.logf(Error, format, args...)
}

func (l *Logger) SaveRawPayloads(rawRequest string, rawResponse string) (requestPath string, responsePath string) {
	if l == nil {
		return "", ""
	}
	if rawRequest != "" {
		path, err := l.SaveText("mineru_request.json", rawRequest)
		if err != nil {
			l.Errorf("failed to persist MinerU request: %s", err)
		} else if path != "" {
			requestPath = path
			l.Verbosef("wrote MinerU raw request to %s", path)
		}
	}
	if rawResponse != "" {
		path, err := l.SaveText("mineru_response.json", rawResponse)
		if err != nil {
			l.Errorf("failed to persist MinerU response: %s", err)
		} else if path != "" {
			responsePath = path
			l.Verbosef("wrote MinerU raw response to %s", path)
		}
	}
	return requestPath, responsePath
}

func (l *Logger) SaveFailure(source any, adapterName string, err error) string {
	if l == nil || err == nil {
		return ""
	}
	artifact := failureArtifact{
		Time:       time.Now().Format(time.RFC3339),
		TaskID:     l.taskID,
		Source:     source,
		AdapterKey: l.adapterKey,
		Error:      err.Error(),
	}
	if adapterName != "" {
		artifact.Adapter = adapterName
	}
	data, marshalErr := json.MarshalIndent(artifact, "", "  ")
	if marshalErr != nil {
		l.Errorf("failed to marshal failure.json: %s", marshalErr)
		return ""
	}
	path, saveErr := l.SaveText("failure.json", string(append(data, '\n')))
	if saveErr != nil {
		l.Errorf("failed to persist failure.json: %s", saveErr)
		return ""
	}
	l.Errorf("wrote failure summary to %s", path)
	return path
}

func (l *Logger) logf(level Level, format string, args ...any) {
	if l == nil {
		return
	}
	message := fmt.Sprintf(format, args...)
	line := l.Write(level, message)
	if line == "" {
		line = formatLine(time.Now(), l.taskID, level, message)
	}
	if l.terminal == nil {
		return
	}
	_, _ = fmt.Fprintf(l.terminal, "%s\n", ColorizeLine(level, line))
}

func ColorizeLine(level Level, line string) string {
	switch NormalizeLevel(level) {
	case Warning:
		return "\033[33m" + line + "\033[0m"
	case Error:
		return "\033[31m" + line + "\033[0m"
	default:
		return line
	}
}
