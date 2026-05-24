package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	tasklogger "github.com/deb-sig/bill-file-converter/core/logger"
)

func removeInputPDFImages(ctx context.Context, input Input, outputDir string, runLogger *tasklogger.Logger) (Input, error) {
	if len(input.Files) == 0 {
		return Input{}, fmt.Errorf("missing input PDF")
	}
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return Input{}, err
	}
	processed := make([]InputFile, 0, len(input.Files))
	for index, file := range input.Files {
		outPath := filepath.Join(outputDir, fmt.Sprintf("input-%03d.pdf", index+1))
		if err := removePDFImages(ctx, file, outPath); err != nil {
			return Input{}, fmt.Errorf("remove images from input %d (%s): %w", index+1, file.Name(), err)
		}
		runLogger.Verbosef("wrote image-free PDF for %s to %s", file.Name(), outPath)
		processed = append(processed, InputFile{
			Path:     outPath,
			FileName: file.Name(),
			MIMEType: "application/pdf",
		})
	}
	return Input{Files: processed}, nil
}

func removePDFImages(ctx context.Context, file InputFile, outputPath string) error {
	gsPath, err := exec.LookPath("gs")
	if err != nil {
		return fmt.Errorf("Ghostscript executable \"gs\" not found; install ghostscript to use this profile")
	}
	inputPath, cleanup, err := materializePDFInput(file, filepath.Dir(outputPath))
	if err != nil {
		return err
	}
	defer cleanup()
	cmd := exec.CommandContext(ctx, gsPath,
		"-q",
		"-dNOPAUSE",
		"-dBATCH",
		"-sDEVICE=pdfwrite",
		"-dFILTERIMAGE",
		"-sOutputFile="+outputPath,
		inputPath,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("run ghostscript: %w: %s", err, strings.TrimSpace(string(output)))
	}
	return nil
}

func materializePDFInput(file InputFile, outputDir string) (string, func(), error) {
	if file.Path != "" {
		return file.Path, func() {}, nil
	}
	tmp, err := os.CreateTemp(outputDir, "source-*.pdf")
	if err != nil {
		return "", nil, err
	}
	if _, err := io.Copy(tmp, file.Reader); err != nil {
		_ = tmp.Close()
		_ = os.Remove(tmp.Name())
		return "", nil, err
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", nil, err
	}
	return tmp.Name(), func() { _ = os.Remove(tmp.Name()) }, nil
}
