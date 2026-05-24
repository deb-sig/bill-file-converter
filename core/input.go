package core

import (
	"fmt"
	"path/filepath"
	"strings"
)

func validatePDFInput(input Input) error {
	if len(input.Files) == 0 {
		return fmt.Errorf("missing input PDF")
	}
	for index, file := range input.Files {
		if err := validatePDFInputFile(file); err != nil {
			if len(input.Files) == 1 {
				return err
			}
			return fmt.Errorf("input %d: %w", index+1, err)
		}
	}
	return nil
}

func validatePDFInputFile(file InputFile) error {
	if file.MIMEType != "" && file.MIMEType != "application/pdf" {
		return fmt.Errorf("unsupported input MIME type %q: only PDF input is supported", file.MIMEType)
	}
	name := file.Name()
	if name != "" && strings.ToLower(filepath.Ext(name)) != ".pdf" {
		return fmt.Errorf("unsupported input file %q: only PDF input is supported", name)
	}
	if file.Path == "" && file.Reader == nil {
		return fmt.Errorf("missing input PDF")
	}
	return nil
}

func sourceInfo(input Input) SourceInfo {
	source := SourceInfo{Files: make([]SourceFileInfo, 0, len(input.Files))}
	for _, file := range input.Files {
		source.Files = append(source.Files, sourceFileInfo(file))
	}
	return source
}

func sourceFileInfo(input InputFile) SourceFileInfo {
	return SourceFileInfo{
		Path:     input.Path,
		FileName: input.Name(),
		MIMEType: input.MIMEType,
	}
}
