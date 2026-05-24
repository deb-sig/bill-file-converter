package core

import (
	"fmt"

	"github.com/deb-sig/bill-file-converter/core/adapters"
)

func ValidateDocument(doc Document, adapter adapters.Adapter) ValidationReport {
	var report ValidationReport
	if len(doc.Tables) == 0 {
		report.Errors = append(report.Errors, "no tables were extracted")
	}

	for tableIdx, table := range doc.Tables {
		if len(table.Headers) == 0 {
			report.Errors = append(report.Errors, fmt.Sprintf("table %d has no headers", tableIdx+1))
		}
		if !equalExactStrings(table.Headers, adapter.Headers) {
			report.Errors = append(report.Errors, fmt.Sprintf("table %d headers do not match adapter profile: got %q, expected %q", tableIdx+1, table.Headers, adapter.Headers))
		}
		for rowIdx, row := range table.Rows {
			if len(row) != len(table.Headers) {
				report.Errors = append(report.Errors, fmt.Sprintf("table %d row %d has %d cells, expected %d", tableIdx+1, rowIdx+1, len(row), len(table.Headers)))
			}
		}
		if len(table.SourcePages) == 0 {
			report.Warnings = append(report.Warnings, fmt.Sprintf("table %d has no source_pages", tableIdx+1))
		}
		report.Warnings = append(report.Warnings, table.Warnings...)
	}

	return report
}

type ValidationError struct {
	Report ValidationReport
}

func (e ValidationError) Error() string {
	return "validation failed"
}

func equalExactStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
