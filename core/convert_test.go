package core

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deb-sig/bill-file-converter/core/adapters"
	tasklogger "github.com/deb-sig/bill-file-converter/core/logger"
)

type fakeMinerU struct {
	result  MinerUParseResult
	err     error
	input   InputFile
	inputs  []InputFile
	results []MinerUParseResult
}

func (m *fakeMinerU) Parse(_ context.Context, file InputFile) (MinerUParseResult, error) {
	m.input = file
	m.inputs = append(m.inputs, file)
	if len(m.results) > 0 {
		result := m.results[0]
		m.results = m.results[1:]
		return result, m.err
	}
	return m.result, m.err
}

func TestConvertUsesMinerUContentListAndWritesArtifacts(t *testing.T) {
	dir := t.TempDir()
	pdf := filepath.Join(dir, "bill.pdf")
	if err := os.WriteFile(pdf, []byte("%PDF-1.7"), 0o644); err != nil {
		t.Fatal(err)
	}
	page := 0
	minerU := &fakeMinerU{result: MinerUParseResult{
		RawRequest:  `{"request":true}`,
		RawResponse: `{"response":true}`,
		ContentList: []MinerUContent{
			{Type: "text", Text: " 招商银行交易流水 \n 招商银行交易流水 "},
			{Type: "text", Text: "招商银行交易流水"},
			{Type: "table", PageIndex: &page, TableBody: `<table>
				<tr><th>记账日期</th><th>货币</th><th>交易金额</th><th>联机余额</th><th>交易摘要</th><th>对手信息</th><th>客户摘要</th></tr>
				<tr><td>Date</td><td>Currency</td><td>Transaction Amount</td><td>Balance</td><td>Transaction Type</td><td>Counter Party</td><td>Customer Summary</td></tr>
				<tr><td>2026-01-01</td><td>CNY</td><td>1.00</td><td>9.00</td><td>消费</td><td></td><td>买菜</td></tr>
				<tr><td>记账日期</td><td>货币</td><td>交易金额</td><td>联机余额</td><td>交易摘要</td><td>对手信息</td><td>客户摘要</td></tr>
				<tr><td>Date</td><td>Currency</td><td>Transaction Amount</td><td>Balance</td><td>Transaction Type</td><td>Counter Party</td><td>Customer Summary</td></tr>
				<tr><td>Date</td><td>Currency</td><td>Transaction Amount</td><td>Balance</td><td>Transaction Ty</td><td></td><td></td></tr>
				<tr><td>Date</td><td>Currency</td><td>Transaction Amount</td><td>Balance</td><td>Transaction Type</td><td>招商银行股份有限公司</td><td></td></tr>
				<tr><td>2026-01-02</td><td>CNY</td><td>-2.00</td><td>7.00</td><td>退款</td><td>张三</td><td>退货</td></tr>
			</table>`},
		},
	}}
	result, err := Convert(context.Background(), Input{Files: []InputFile{{Path: pdf, FileName: "bill.pdf"}}}, Options{
		MinerU:          minerU,
		AdapterKey:      "cmb_debit",
		AdapterRegistry: testRegistry(),
		OutputDir:       filepath.Join(dir, "output"),
	})
	if err != nil {
		t.Fatalf("%v: %#v", err, result.ValidationReport)
	}
	if result.Metadata["raw_text"] != "招商银行交易流水 招商银行交易流水\n招商银行交易流水" {
		t.Fatalf("unexpected raw_text: %#v", result.Metadata)
	}
	if len(result.Tables) != 1 || len(result.Tables[0].Rows) != 2 {
		t.Fatalf("unexpected tables: %#v", result.Tables)
	}
	if result.Tables[0].SourcePages[0] != 1 {
		t.Fatalf("expected source page 1, got %#v", result.Tables[0].SourcePages)
	}
	for _, path := range []string{
		result.Artifacts.JSONPath,
		result.Artifacts.CSVPath,
		result.Artifacts.ContentListPath,
		result.Artifacts.ProcessLogPath,
		filepath.Join(result.Artifacts.LoggerDir, "mineru_request.json"),
		filepath.Join(result.Artifacts.LoggerDir, "mineru_response.json"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected artifact %s: %v", path, err)
		}
	}
	csvData, err := os.ReadFile(result.Artifacts.CSVPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(csvData), "记账日期,货币,交易金额,联机余额,交易摘要,对手信息,客户摘要") != 1 {
		t.Fatalf("expected repeated header to be removed: %s", csvData)
	}
	if strings.Contains(string(csvData), "Date,Currency,Transaction Amount,Balance,Transaction Type,Counter Party,Customer Summary") {
		t.Fatalf("expected English header alias to be removed: %s", csvData)
	}
	if strings.Contains(string(csvData), "Transaction Ty") {
		t.Fatalf("expected truncated English header alias to be removed: %s", csvData)
	}
	if strings.Contains(string(csvData), "招商银行股份有限公司") {
		t.Fatalf("expected first-column Date header row to be removed: %s", csvData)
	}
}

func TestConvertMultiplePDFs(t *testing.T) {
	dir := t.TempDir()
	first := filepath.Join(dir, "1.pdf")
	second := filepath.Join(dir, "2.pdf")
	for _, path := range []string{first, second} {
		if err := os.WriteFile(path, []byte("%PDF-1.7"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	minerU := &fakeMinerU{results: []MinerUParseResult{
		{ContentList: []MinerUContent{{Type: "table", TableBody: `<table>
				<tr><th>交易日期</th><th>交易时间</th><th>交易摘要</th><th>交易金额</th><th>本次余额</th><th>对手信息</th><th>日 志 号</th><th>交易渠道</th><th>交易附言</th></tr>
				<tr><td>20260101</td><td>12:00:00</td><td>转账</td><td>1.00</td><td>2.00</td><td>--</td><td>1234567890</td><td>网银</td><td></td></tr>
			</table>`}}},
		{ContentList: []MinerUContent{{Type: "table", TableBody: `<table>
				<tr><th>交易日期</th><th>交易时间</th><th>交易摘要</th><th>交易金额</th><th>本次余额</th><th>对手信息</th><th>日 志 号</th><th>交易渠道</th><th>交易附言</th></tr>
				<tr><td>20260102</td><td>13:00:00</td><td>转账</td><td>3.00</td><td>5.00</td><td>--</td><td>1234567891</td><td>网银</td><td></td></tr>
			</table>`}}},
	}}
	result, err := Convert(context.Background(), Input{Files: []InputFile{
		{Path: first, FileName: "1.pdf"},
		{Path: second, FileName: "2.pdf"},
	}}, Options{
		MinerU:          minerU,
		AdapterKey:      "abc_debit",
		AdapterRegistry: testRegistry(),
		OutputDir:       filepath.Join(dir, "output"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(minerU.inputs) != 2 || minerU.inputs[0].FileName != "1.pdf" || minerU.inputs[1].FileName != "2.pdf" {
		t.Fatalf("expected sequential input files, got %#v", minerU.inputs)
	}
	if got := result.Tables[0].Headers[6]; got != "日志号" {
		t.Fatalf("expected header spaces to be stripped via profile headers, got %q", got)
	}
	if len(result.Tables[0].Rows) != 2 {
		t.Fatalf("expected merged rows from two PDFs, got %#v", result.Tables[0].Rows)
	}
}

func TestCmbCreditSkipsSectionRowsAndSummaryRows(t *testing.T) {
	dir := t.TempDir()
	pdf := filepath.Join(dir, "bill.pdf")
	if err := os.WriteFile(pdf, []byte("%PDF-1.7"), 0o644); err != nil {
		t.Fatal(err)
	}
	minerU := &fakeMinerU{result: MinerUParseResult{
		ContentList: []MinerUContent{
			{Type: "text", Text: "招商银行信用卡对账单（个人消费卡账户 2025年02月）（补）"},
			{Type: "table", TableBody: `<table>
				<tr><th>交易日 SOLD</th><th>记账日 POSTED</th><th>交易摘要 DESCRIPTION</th><th>人民币金额 RMB AMOUNT</th><th>卡号末四位 CARD NO(Last 4digits)</th><th>交易地金额 Original Tran Amount</th></tr>
				<tr><td>还款</td><td></td><td></td><td></td><td></td><td></td></tr>
				<tr><td>还款</td><td>还款</td><td>还款</td><td>还款</td><td>还款</td><td>还款</td></tr>
				<tr><td></td><td>01/27</td><td>银联转账还款</td><td>-29.50</td><td>2508</td><td>-29.50</td></tr>
				<tr><td>消费</td><td></td><td></td><td></td><td></td><td></td></tr>
				<tr><td>01/12</td><td>01/13</td><td>支付宝-上海拉扎斯信息科技有限公司</td><td>1.70</td><td>2508</td><td>1.70(CN)</td></tr>
				<tr><td>本期还款总额Current Balance</td><td>=</td><td>上期账单金额Balance B/F</td><td>-</td><td>上期还款金额Payment</td><td>+</td></tr>
			</table>`},
		},
	}}
	result, err := Convert(context.Background(), Input{Files: []InputFile{{Path: pdf, FileName: "bill.pdf"}}}, Options{
		MinerU:          minerU,
		AdapterKey:      "cmb_credit",
		AdapterRegistry: adapters.BuiltinRegistry(),
		OutputDir:       filepath.Join(dir, "output"),
	})
	if err != nil {
		t.Fatal(err)
	}
	csvData, err := os.ReadFile(result.Artifacts.CSVPath)
	if err != nil {
		t.Fatal(err)
	}
	got := string(csvData)
	if !strings.HasPrefix(got, "交易日,记账日,交易摘要,人民币金额,卡号末四位,交易地金额\n") {
		t.Fatalf("unexpected csv prefix: %s", got)
	}
	if strings.Contains(got, "\n还款,") || strings.Contains(got, "\n消费,") {
		t.Fatalf("expected section rows to be skipped: %s", got)
	}
	if strings.Contains(got, "本期还款总额") {
		t.Fatalf("expected summary rows to be skipped: %s", got)
	}
}

func TestConvertValidationFailsWithoutMatchingTables(t *testing.T) {
	dir := t.TempDir()
	pdf := filepath.Join(dir, "bill.pdf")
	if err := os.WriteFile(pdf, []byte("%PDF-1.7"), 0o644); err != nil {
		t.Fatal(err)
	}
	minerU := &fakeMinerU{result: MinerUParseResult{
		ContentList: []MinerUContent{{Type: "table", TableBody: `<table><tr><th>A</th></tr><tr><td>B</td></tr></table>`}},
	}}
	result, err := Convert(context.Background(), Input{Files: []InputFile{{Path: pdf, FileName: "bill.pdf"}}}, Options{
		MinerU:          minerU,
		AdapterKey:      "cmb_debit",
		AdapterRegistry: testRegistry(),
		OutputDir:       filepath.Join(dir, "output"),
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
	if result.Artifacts.JSONPath == "" {
		t.Fatal("expected result JSON to be written before validation failure")
	}
}

func TestValueMatchesGuardFormatStrictDates(t *testing.T) {
	cases := []struct {
		value  string
		format adapters.RowGuardFormat
		want   bool
	}{
		{value: "2024-02-29", format: adapters.RowGuardFormatYYYYDashMMDashDD, want: true},
		{value: "2026-01-01", format: adapters.RowGuardFormatYYYYDashMMDashDD, want: true},
		{value: "2025-02-29", format: adapters.RowGuardFormatYYYYDashMMDashDD, want: false},
		{value: "2025-2-09", format: adapters.RowGuardFormatYYYYDashMMDashDD, want: false},
		{value: "20240229", format: adapters.RowGuardFormatYYYYMMDD, want: true},
		{value: "20250229", format: adapters.RowGuardFormatYYYYMMDD, want: false},
		{value: "2025029", format: adapters.RowGuardFormatYYYYMMDD, want: false},
		{value: "2024022923:59:59", format: adapters.RowGuardFormatYYYYMMDDHHMMSS, want: true},
		{value: "2025022923:59:59", format: adapters.RowGuardFormatYYYYMMDDHHMMSS, want: false},
		{value: "2024022924:00:00", format: adapters.RowGuardFormatYYYYMMDDHHMMSS, want: false},
		{value: "2024-10-25 00:22:21", format: adapters.RowGuardFormatYYYYDashMMDashDDHHMMSS, want: true},
		{value: "2024-10-25 24:00:00", format: adapters.RowGuardFormatYYYYDashMMDashDDHHMMSS, want: false},
		{value: "2024-10-25T00:22:21", format: adapters.RowGuardFormatYYYYDashMMDashDDHHMMSS, want: false},
		{value: "02/29", format: adapters.RowGuardFormatMMSlashDD, want: true},
		{value: "02/31", format: adapters.RowGuardFormatMMSlashDD, want: false},
		{value: "13/01", format: adapters.RowGuardFormatMMSlashDD, want: false},
		{value: "2/09", format: adapters.RowGuardFormatMMSlashDD, want: false},
		{value: "1", format: adapters.RowGuardFormatPositiveInteger, want: true},
		{value: "001", format: adapters.RowGuardFormatPositiveInteger, want: true},
		{value: "1.0", format: adapters.RowGuardFormatPositiveInteger, want: false},
		{value: "打印完毕", format: adapters.RowGuardFormatPositiveInteger, want: false},
		{value: "02/09", format: adapters.RowGuardFormat("unknown"), want: false},
	}
	for _, tc := range cases {
		if got := valueMatchesGuardFormat(tc.value, tc.format); got != tc.want {
			t.Fatalf("valueMatchesGuardFormat(%q, %q) = %v, want %v", tc.value, tc.format, got, tc.want)
		}
	}
}

func TestParseHTMLTableSpans(t *testing.T) {
	rows, err := parseHTMLTable(`<table>
		<tr><th rowspan="2">A</th><th colspan="2">B</th></tr>
		<tr><th>C</th><th>D</th></tr>
		<tr><td>1</td><td></td><td>3</td></tr>
	</table>`)
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"A", "B", "B"}, {"A", "C", "D"}, {"1", "", "3"}}
	if len(rows) != len(want) {
		t.Fatalf("rows=%#v", rows)
	}
	for i := range want {
		if strings.Join(rows[i], "|") != strings.Join(want[i], "|") {
			t.Fatalf("row %d = %#v, want %#v", i, rows[i], want[i])
		}
	}
}

func TestColorizeStdLogLine(t *testing.T) {
	if got := tasklogger.ColorizeLine(tasklogger.Warning, "warn"); got != "\033[33mwarn\033[0m" {
		t.Fatalf("warning color = %q", got)
	}
	if got := tasklogger.ColorizeLine(tasklogger.Error, "err"); got != "\033[31merr\033[0m" {
		t.Fatalf("error color = %q", got)
	}
	if got := tasklogger.ColorizeLine(tasklogger.Info, "info"); got != "info" {
		t.Fatalf("info color = %q", got)
	}
}

type mapAdapterRegistry map[string]adapters.Adapter

func (r mapAdapterRegistry) MustGet(key string) (adapters.Adapter, error) {
	adapter, ok := r[key]
	if !ok {
		return adapters.Adapter{}, fmt.Errorf("missing test adapter %q", key)
	}
	return adapter, nil
}

func testRegistry() mapAdapterRegistry {
	cmb, err := adapters.BuiltinRegistry().MustGet("cmb_debit")
	if err != nil {
		panic(err)
	}
	cmb.RemoveImages = false
	abc, err := adapters.BuiltinRegistry().MustGet("abc_debit")
	if err != nil {
		panic(err)
	}
	abc.RemoveImages = false
	return mapAdapterRegistry{
		cmb.Key: cmb,
		abc.Key: abc,
	}
}
