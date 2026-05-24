package core

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/deb-sig/bill-file-converter/core/adapters"
)

func DocumentFromMinerUContent(contentList []MinerUContent, adapter adapters.Adapter) Document {
	doc := Document{
		Metadata: map[string]string{},
	}
	rawText := dedupeRawText(contentList)
	if rawText != "" {
		doc.Metadata["raw_text"] = rawText
	}

	for _, item := range contentList {
		if strings.ToLower(item.Type) != "table" || strings.TrimSpace(item.TableBody) == "" {
			continue
		}
		rows, err := parseHTMLTableCells(item.TableBody)
		if err != nil {
			doc.Tables = append(doc.Tables, Table{Warnings: []string{fmt.Sprintf("parse table html: %s", err)}})
			continue
		}
		table, ok := tableFromRows(rows, adapter, pageFromContent(item))
		if !ok {
			continue
		}
		doc = mergeCleanedTable(doc, table)
	}
	return doc
}

func dedupeRawText(contentList []MinerUContent) string {
	seen := map[string]bool{}
	values := []string{}
	for _, item := range contentList {
		if strings.ToLower(item.Type) == "table" {
			continue
		}
		text := normalizeText(item.Text)
		if text == "" || seen[text] {
			continue
		}
		seen[text] = true
		values = append(values, text)
	}
	return strings.Join(values, "\n")
}

func tableFromRows(rows [][]tableCell, adapter adapters.Adapter, page int) (Table, bool) {
	headerIndex, headers := matchingHeaderRow(rows, adapter)
	if headerIndex < 0 {
		return Table{}, false
	}
	table := Table{
		Headers:     headers,
		SourcePages: mergeSortedUniqueInts(nil, page),
	}
	for _, row := range rows[headerIndex+1:] {
		if stringRowIsEmpty(tableCellTexts(row)) || rowLooksLikeHeader(tableCellTexts(row), adapter, headers) {
			continue
		}
		normalized := normalizeRowWidth(row, len(headers))
		if !rowMatchesGuards(tableCellTexts(normalized), adapter) {
			continue
		}
		cells := make([]*string, len(normalized))
		for i, cell := range normalized {
			if cell.RowspanCarryover && adapterBlanksRowspanCarryoverColumn(adapter, i) {
				continue
			}
			value := normalizeText(cell.Text)
			if value == "" {
				continue
			}
			cells[i] = &value
		}
		table.Rows = append(table.Rows, cells)
	}
	if len(table.Rows) == 0 {
		return Table{}, false
	}
	return table, true
}

func matchingHeaderRow(rows [][]tableCell, adapter adapters.Adapter) (int, []string) {
	for rowIndex, row := range rows {
		normalized := normalizeStringRow(tableCellTexts(row))
		if rowMatchesHeaderCandidate(normalized, adapter.Headers) {
			return rowIndex, append([]string(nil), adapter.Headers...)
		}
		for _, alias := range adapter.HeaderAliases {
			if rowMatchesHeaderCandidate(normalized, alias) {
				return rowIndex, append([]string(nil), adapter.Headers...)
			}
		}
	}
	return -1, nil
}

func adapterBlanksRowspanCarryoverColumn(adapter adapters.Adapter, column int) bool {
	for _, candidate := range adapter.BlankRowspanCarryoverColumns {
		if candidate == column {
			return true
		}
	}
	return false
}

func rowLooksLikeHeader(row []string, adapter adapters.Adapter, headers []string) bool {
	normalized := normalizeStringRow(row)
	if rowMatchesHeaderCandidate(normalized, headers) || rowHasHeaderFirstCell(normalized, headers) {
		return true
	}
	for _, alias := range adapter.HeaderAliases {
		if rowMatchesHeaderCandidate(normalized, alias) || rowLooksLikeTruncatedHeader(normalized, alias) || rowHasHeaderFirstCell(normalized, alias) {
			return true
		}
	}
	return false
}

func rowMatchesHeaderCandidate(row []string, header []string) bool {
	return equalNormalizedStrings(row, header) || equalCollapsedHeaders(row, header)
}

func rowHasHeaderFirstCell(row []string, header []string) bool {
	if len(row) == 0 || len(header) == 0 {
		return false
	}
	first := normalizeText(row[0])
	return first != "" && first == normalizeText(header[0])
}

func rowLooksLikeTruncatedHeader(row []string, header []string) bool {
	if len(row) != len(header) {
		return false
	}
	matches := 0
	for i := range header {
		cell := normalizeText(row[i])
		expected := normalizeText(header[i])
		if cell == "" {
			continue
		}
		if cell == expected || strings.HasPrefix(expected, cell) {
			matches++
			continue
		}
		return false
	}
	required := (len(header) + 1) / 2
	if required < 3 {
		required = len(header)
	}
	return matches >= required
}

func normalizeStringRow(row []string) []string {
	normalized := make([]string, len(row))
	for i, value := range row {
		normalized[i] = normalizeText(value)
	}
	return normalized
}

func normalizeRowWidth(row []tableCell, width int) []tableCell {
	normalized := make([]tableCell, len(row))
	for i, cell := range row {
		normalized[i] = tableCell{
			Text:             normalizeText(cell.Text),
			RowspanCarryover: cell.RowspanCarryover,
		}
	}
	if len(normalized) > width {
		return normalized[:width]
	}
	for len(normalized) < width {
		normalized = append(normalized, tableCell{})
	}
	return normalized
}

func rowMatchesGuards(row []string, adapter adapters.Adapter) bool {
	for _, guard := range adapter.RowGuards {
		if guard.Column < 0 || guard.Column >= len(row) {
			return false
		}
		if !valueMatchesGuardFormat(normalizeText(row[guard.Column]), guard.Format) {
			return false
		}
	}
	return true
}

func valueMatchesGuardFormat(value string, format adapters.RowGuardFormat) bool {
	switch format {
	case adapters.RowGuardFormatYYYYDashMMDashDD:
		parsed, err := time.Parse("2006-01-02", value)
		return err == nil && parsed.Format("2006-01-02") == value
	case adapters.RowGuardFormatYYYYMMDD:
		parsed, err := time.Parse("20060102", value)
		return err == nil && parsed.Format("20060102") == value
	case adapters.RowGuardFormatYYYYMMDDHHMMSS:
		parsed, err := time.Parse("2006010215:04:05", value)
		return err == nil && parsed.Format("2006010215:04:05") == value
	case adapters.RowGuardFormatYYYYDashMMDashDDHHMMSS:
		parsed, err := time.Parse("2006-01-02 15:04:05", value)
		return err == nil && parsed.Format("2006-01-02 15:04:05") == value
	case adapters.RowGuardFormatMMSlashDD:
		parsed, err := time.Parse("2006/01/02", "2000/"+value)
		return err == nil && parsed.Format("01/02") == value
	case adapters.RowGuardFormatPositiveInteger:
		if value == "" {
			return false
		}
		for _, ch := range value {
			if ch < '0' || ch > '9' {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func stringRowIsEmpty(row []string) bool {
	for _, value := range row {
		if normalizeText(value) != "" {
			return false
		}
	}
	return true
}

func mergeCleanedTable(doc Document, table Table) Document {
	targetIndex := matchingTableIndex(doc.Tables, table)
	if targetIndex < 0 {
		doc.Tables = append(doc.Tables, table)
		return doc
	}
	target := &doc.Tables[targetIndex]
	target.Rows = append(target.Rows, table.Rows...)
	target.SourcePages = mergeSortedUniqueInts(target.SourcePages, table.SourcePages...)
	target.Warnings = append(target.Warnings, table.Warnings...)
	return doc
}

func matchingTableIndex(tables []Table, candidate Table) int {
	for i, table := range tables {
		if equalNormalizedStrings(table.Headers, candidate.Headers) {
			return i
		}
	}
	return -1
}

func equalNormalizedStrings(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if normalizeText(a[i]) != normalizeText(b[i]) {
			return false
		}
	}
	return true
}

func equalCollapsedHeaders(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if normalizeHeaderText(a[i]) != normalizeHeaderText(b[i]) {
			return false
		}
	}
	return true
}

func normalizeHeaderText(value string) string {
	return strings.ReplaceAll(normalizeText(value), " ", "")
}

func mergeSortedUniqueInts(base []int, values ...int) []int {
	seen := make(map[int]bool, len(base)+len(values))
	for _, value := range base {
		seen[value] = true
	}
	for _, value := range values {
		if value == 0 || seen[value] {
			continue
		}
		base = append(base, value)
		seen[value] = true
	}
	sort.Ints(base)
	return base
}

func pageFromContent(item MinerUContent) int {
	if item.PageIndex == nil {
		return 0
	}
	return *item.PageIndex + 1
}

func normalizeText(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
