package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	tasklogger "github.com/deb-sig/bill-file-converter/core/logger"
)

func Convert(ctx context.Context, input Input, options Options) (Result, error) {
	taskID, err := newTaskID(time.Now())
	if err != nil {
		return Result{}, err
	}
	baseOutputDir := options.OutputDir
	if baseOutputDir == "" {
		baseOutputDir = "."
	}
	options.OutputDir = filepath.Join(baseOutputDir, taskID)
	resultDir := filepath.Join(options.OutputDir, "result")
	for _, dir := range []string{resultDir, options.OutputDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return Result{}, err
		}
	}
	runLogger, err := tasklogger.NewTaskLogger(options.OutputDir, taskID, options.AdapterKey, options.LogWriter)
	if err != nil {
		return Result{}, err
	}
	defer runLogger.Close()
	runLogger.Infof("checking input")
	if err := validatePDFInput(input); err != nil {
		runLogger.Errorf("input validation failed: %s", err)
		runLogger.SaveFailure(sourceInfo(input), "", err)
		return Result{}, err
	}
	registry := options.AdapterRegistry
	if registry == nil {
		err := fmt.Errorf("missing adapter registry")
		runLogger.Errorf("adapter registry missing")
		runLogger.SaveFailure(sourceInfo(input), "", err)
		return Result{}, err
	}
	adapter, err := registry.MustGet(options.AdapterKey)
	if err != nil {
		runLogger.Errorf("adapter lookup failed: %s", err)
		runLogger.SaveFailure(sourceInfo(input), "", err)
		return Result{}, err
	}
	runLogger.Infof("using adapter %s (%s)", adapter.Key, adapter.Name)
	if options.MinerU == nil {
		err := fmt.Errorf("missing MinerU client")
		runLogger.Errorf("MinerU client missing")
		runLogger.SaveFailure(sourceInfo(input), adapter.Name, err)
		return Result{}, err
	}

	parseInput := input
	if adapter.RemoveImages {
		runLogger.Infof("removing images from PDF input(s) before MinerU parsing")
		var preprocessErr error
		parseInput, preprocessErr = removeInputPDFImages(ctx, input, filepath.Join(runLogger.Dir(), "preprocessed"), runLogger)
		if preprocessErr != nil {
			runLogger.Errorf("PDF image removal failed: %s", preprocessErr)
			runLogger.SaveFailure(sourceInfo(input), adapter.Name, preprocessErr)
			return Result{}, preprocessErr
		}
	}

	runLogger.Infof("submitting %d PDF input(s) to MinerU", len(parseInput.Files))
	parseResult, err := parseMinerUInInputOrder(ctx, options.MinerU, parseInput.Files)
	rawRequestPath, rawResponsePath := runLogger.SaveRawPayloads(parseResult.RawRequest, parseResult.RawResponse)
	if err != nil {
		runLogger.Errorf("MinerU parse failed: %s", err)
		runLogger.SaveFailure(sourceInfo(input), adapter.Name, err)
		return Result{}, err
	}
	contentListPath := filepath.Join(runLogger.Dir(), "content_list.json")
	if err := writeContentList(contentListPath, parseResult.ContentList); err != nil {
		runLogger.Errorf("writing content_list.json failed: %s", err)
		runLogger.SaveFailure(sourceInfo(input), adapter.Name, err)
		return Result{}, err
	}

	runLogger.Infof("cleaning MinerU content list")
	doc := DocumentFromMinerUContent(parseResult.ContentList, adapter)
	runLogger.Infof("validating extracted document")
	report := ValidateDocument(doc, adapter)
	var csvBytes []byte
	if !options.SkipCSV {
		runLogger.Infof("exporting CSV")
		var csvErr error
		csvBytes, csvErr = ExportCSV(doc)
		if csvErr != nil {
			report.Errors = append(report.Errors, csvErr.Error())
			runLogger.Errorf("CSV export failed: %s", csvErr)
		}
	} else {
		runLogger.Infof("skipping CSV export")
	}

	result := Result{
		TaskID:           taskID,
		AdapterKey:       adapter.Key,
		AdapterName:      adapter.Name,
		Source:           sourceInfo(input),
		GeneratedAt:      time.Now().Format(time.RFC3339),
		Metadata:         doc.Metadata,
		Tables:           doc.Tables,
		ValidationReport: report,
		Artifacts: Artifacts{
			LoggerDir:             runLogger.Dir(),
			ProcessLogPath:        runLogger.Path(),
			ContentListPath:       contentListPath,
			MinerURawRequestPath:  rawRequestPath,
			MinerURawResponsePath: rawResponsePath,
		},
	}

	runLogger.Infof("writing artifacts to %s", options.OutputDir)
	if err := writeArtifacts(resultDir, &result, csvBytes); err != nil {
		runLogger.Errorf("writing result artifacts failed: %s", err)
		result.Artifacts.FailurePath = runLogger.SaveFailure(sourceInfo(input), adapter.Name, err)
		return Result{}, err
	}
	if report.HasErrors() {
		runLogger.Errorf("validation failed with %d error(s)", len(report.Errors))
		for _, validationErr := range report.Errors {
			runLogger.Errorf("validation error: %s", validationErr)
		}
		result.Artifacts.FailurePath = runLogger.SaveFailure(sourceInfo(input), adapter.Name, ValidationError{Report: report})
		return result, ValidationError{Report: report}
	}
	runLogger.Infof("done")
	return result, nil
}
