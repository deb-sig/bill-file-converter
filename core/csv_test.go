package core

import (
	"strings"
	"testing"
)

func TestExportCSVMultipleTablesAndNullCells(t *testing.T) {
	value := "abc"
	data, err := ExportCSV(Document{
		Tables: []Table{
			{Headers: []string{"A", "B"}, Rows: [][]*string{{&value, nil}}},
			{Name: "Second", Headers: []string{"C"}, Rows: [][]*string{{nil}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "Second") {
		t.Fatalf("csv should not contain table names: %q", text)
	}
	for _, want := range []string{"A,B", "abc,", "C"} {
		if !strings.Contains(text, want) {
			t.Fatalf("csv missing %q: %q", want, text)
		}
	}
}

func TestExportCSVDeduplicatesRepeatedHeadersWithoutBlankLines(t *testing.T) {
	one := "1"
	two := "2"
	data, err := ExportCSV(Document{
		Tables: []Table{
			{Headers: []string{"A", "B"}, Rows: [][]*string{{&one, nil}}},
			{Headers: []string{"A", "B"}, Rows: [][]*string{{&two, nil}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Count(text, "A,B") != 1 {
		t.Fatalf("expected one header, got %q", text)
	}
	if strings.Contains(text, "\n\n") {
		t.Fatalf("expected no blank lines, got %q", text)
	}
	want := "A,B\n1,\n2,\n"
	if text != want {
		t.Fatalf("csv = %q, want %q", text, want)
	}
}
