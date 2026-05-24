package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/deb-sig/bill-file-converter/core"
	"github.com/deb-sig/bill-file-converter/core/adapters"
)

func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}
	switch args[0] {
	case "convert":
		return runConvert(ctx, args[1:], stdout, stderr, false)
	case "inspect":
		return runConvert(ctx, args[1:], stdout, stderr, true)
	case "list-types":
		return runListTypes(stdout)
	case "config":
		return runConfig(args[1:], stdout, stderr)
	case "mineru":
		return runMinerU(ctx, args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func runConvert(ctx context.Context, args []string, stdout, stderr io.Writer, inspect bool) int {
	fs := flag.NewFlagSet("convert", flag.ContinueOnError)
	fs.SetOutput(stderr)
	typeKey := fs.String("type", "", "bill type adapter key")
	configPath := fs.String("config", "config.yaml", "config file path")
	outputDir := fs.String("output", "output", "output directory")
	args = normalizeFlagArgs(args, map[string]bool{
		"-type": true, "--type": true,
		"-config": true, "--config": true,
		"-output": true, "--output": true,
	})
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, "expected at least one PDF file or directory")
		return 2
	}
	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}
	minerUConfig, err := config.MinerUHTTPConfig()
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}
	minerU, err := core.NewMinerUHTTPClient(minerUConfig)
	if err != nil {
		fmt.Fprintf(stderr, "create MinerU client: %v\n", err)
		return 1
	}
	inputFiles, err := expandInputPaths(fs.Args())
	if err != nil {
		fmt.Fprintf(stderr, "input: %v\n", err)
		return 1
	}
	input := core.Input{}
	for _, path := range inputFiles {
		input.Files = append(input.Files, core.InputFile{Path: path, FileName: filepath.Base(path)})
	}
	result, err := core.Convert(ctx, input, core.Options{
		MinerU:          minerU,
		AdapterKey:      *typeKey,
		AdapterRegistry: adapterRegistryForCLI(),
		OutputDir:       *outputDir,
		SkipCSV:         inspect,
		LogWriter:       stderr,
	})
	if err != nil {
		var validation core.ValidationError
		if errors.As(err, &validation) {
			fmt.Fprintf(stderr, "validation failed; wrote %s\n", result.Artifacts.JSONPath)
			return 1
		}
		fmt.Fprintf(stderr, "convert: %v\n", err)
		return 1
	}
	if inspect {
		fmt.Fprintf(stdout, "inspection completed: %s\n", result.Artifacts.JSONPath)
		return 0
	}
	fmt.Fprintf(stdout, "task %s wrote %s and %s\n", result.TaskID, result.Artifacts.JSONPath, result.Artifacts.CSVPath)
	return 0
}

func expandInputPaths(paths []string) ([]string, error) {
	var expanded []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			if strings.ToLower(filepath.Ext(path)) != ".pdf" {
				return nil, fmt.Errorf("unsupported input file %q: only PDF input is supported", path)
			}
			expanded = append(expanded, path)
			continue
		}
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		pdfs := []string{}
		for _, entry := range entries {
			if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".pdf" {
				continue
			}
			pdfs = append(pdfs, filepath.Join(path, entry.Name()))
		}
		sort.Slice(pdfs, func(i, j int) bool {
			return naturalLess(filepath.Base(pdfs[i]), filepath.Base(pdfs[j]))
		})
		if len(pdfs) == 0 {
			return nil, fmt.Errorf("directory %q contains no PDF files", path)
		}
		expanded = append(expanded, pdfs...)
	}
	return expanded, nil
}

func naturalLess(a, b string) bool {
	ai, bi := 0, 0
	for ai < len(a) && bi < len(b) {
		ar, br := a[ai], b[bi]
		if isDigit(ar) && isDigit(br) {
			an, nextA := readNumber(a, ai)
			bn, nextB := readNumber(b, bi)
			if an != bn {
				return an < bn
			}
			ai, bi = nextA, nextB
			continue
		}
		al := strings.ToLower(string(ar))
		bl := strings.ToLower(string(br))
		if al != bl {
			return al < bl
		}
		ai++
		bi++
	}
	return len(a) < len(b)
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func readNumber(value string, start int) (int, int) {
	end := start
	for end < len(value) && isDigit(value[end]) {
		end++
	}
	parsed, err := strconv.Atoi(value[start:end])
	if err != nil {
		return 0, end
	}
	return parsed, end
}

func runListTypes(stdout io.Writer) int {
	for _, adapter := range adapterRegistryForCLI().List() {
		fmt.Fprintf(stdout, "%s\t%s\n", adapter.Key, adapter.Name)
	}
	return 0
}

type listedAdapterRegistry interface {
	core.AdapterRegistry
	List() []adapters.Adapter
}

func adapterRegistryForCLI() listedAdapterRegistry {
	registry := adapters.BuiltinRegistry()
	// Test-only escape hatch for CLI e2e tests that run a fake MinerU server
	// with placeholder PDF files. Production runs should leave this unset so
	// profiles that need PDF image removal still exercise Ghostscript.
	if os.Getenv("BFC_E2E_SKIP_REMOVE_IMAGES") != "1" {
		return registry
	}
	values := registry.List()
	for i := range values {
		values[i].RemoveImages = false
	}
	return adapters.NewRegistry(values...)
}

func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "init" {
		fmt.Fprintln(stderr, "usage: config init [-output config.yaml]")
		return 2
	}
	fs := flag.NewFlagSet("config init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	out := fs.String("output", "config.yaml", "config output path")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	if _, err := os.Stat(*out); err == nil {
		fmt.Fprintf(stderr, "%s already exists\n", *out)
		return 1
	}
	if err := WriteDefaultConfig(*out); err != nil {
		fmt.Fprintf(stderr, "write config: %v\n", err)
		return 1
	}
	fmt.Fprintf(stdout, "wrote %s\n", *out)
	return 0
}

func runMinerU(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] != "test" {
		fmt.Fprintln(stderr, "usage: mineru test [-config config.yaml]")
		return 2
	}
	fs := flag.NewFlagSet("mineru test", flag.ContinueOnError)
	fs.SetOutput(stderr)
	configPath := fs.String("config", "config.yaml", "config file path")
	if err := fs.Parse(args[1:]); err != nil {
		return 2
	}
	config, err := LoadConfig(*configPath)
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}
	minerUConfig, err := config.MinerUHTTPConfig()
	if err != nil {
		fmt.Fprintf(stderr, "load config: %v\n", err)
		return 1
	}
	minerU, err := core.NewMinerUHTTPClient(minerUConfig)
	if err != nil {
		fmt.Fprintf(stderr, "create MinerU client: %v\n", err)
		return 1
	}
	if err := minerU.Ping(ctx); err != nil {
		fmt.Fprintf(stderr, "MinerU ping failed: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "MinerU reachable")
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "usage: bill-file-converter <convert|inspect|list-types|config|mineru> [options]")
}

func normalizeFlagArgs(args []string, flagsWithValue map[string]bool) []string {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--" {
			positionals = append(positionals, args[i+1:]...)
			break
		}
		if len(arg) == 0 || arg[0] != '-' || arg == "-" {
			positionals = append(positionals, arg)
			continue
		}
		flags = append(flags, arg)
		name := arg
		for idx, ch := range arg {
			if ch == '=' {
				name = arg[:idx]
				break
			}
		}
		if flagsWithValue[name] && !hasInlineValue(arg) && i+1 < len(args) {
			i++
			flags = append(flags, args[i])
		}
	}

	return append(flags, positionals...)
}

func hasInlineValue(arg string) bool {
	for _, ch := range arg {
		if ch == '=' {
			return true
		}
	}
	return false
}
