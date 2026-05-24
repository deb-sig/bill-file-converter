package core

import (
	"bytes"
	"encoding/csv"
	"fmt"
)

func ExportCSV(doc Document) ([]byte, error) {
	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	var firstHeaders []string

	for _, table := range doc.Tables {
		if len(table.Headers) > 0 {
			if firstHeaders == nil {
				firstHeaders = append([]string(nil), table.Headers...)
				if err := writer.Write(table.Headers); err != nil {
					return nil, err
				}
			} else if !equalExactStrings(table.Headers, firstHeaders) {
				if err := writer.Write(table.Headers); err != nil {
					return nil, err
				}
			}
		}
		for _, row := range table.Rows {
			values := make([]string, len(row))
			for idx, cell := range row {
				if cell != nil {
					values[idx] = *cell
				}
			}
			if err := writer.Write(values); err != nil {
				return nil, err
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("write csv: %w", err)
	}
	return buf.Bytes(), nil
}
