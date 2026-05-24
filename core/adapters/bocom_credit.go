package adapters

func bocomCreditRegularAdapter() Adapter {
	// 正常发送账单的 PDF 表格包含卡末四位，日期是 MM/DD。
	// MinerU 可能把表头识别成中英混合一行，也可能识别成中文拆行后的英文表头行。
	// 表头匹配会折叠空格，因此 alias 中保留可读空格，但仍能匹配 MinerU 的无空格拼接形态。
	headers := []string{"交易日期", "记账日期", "卡末四位", "交易说明", "交易币种/金额", "入账币种/金额"}
	normalPDFHeaders := []string{
		"交易日期 Transaction Date",
		"记账日期 Posting Date",
		"卡末四位 Card Number (Last 4 digits)",
		"交易说明 Description of Transaction",
		"交易金额 Transaction Curr/Amt",
		"入账金额 Payment Curr/Amt",
	}
	normalPDFEnglishHeaders := []string{
		"Transaction Date",
		"Posting Date",
		"Card Number (Last 4 digits)",
		"Description of Transaction",
		"Transaction Curr/Amt",
		"Payment Curr/Amt",
	}
	return Adapter{
		Key:           "bocom_credit_regular",
		Name:          "交通银行信用卡（正常账单）",
		Headers:       headers,
		HeaderAliases: [][]string{normalPDFHeaders, normalPDFEnglishHeaders},
		RowGuards: []RowGuard{
			{Column: 0, Format: RowGuardFormatMMSlashDD},
			{Column: 1, Format: RowGuardFormatMMSlashDD},
		},
	}
}

func bocomCreditReissueAdapter() Adapter {
	// 补发账单的 PDF 表格沿用 legacy EML 的五列结构，没有卡末四位；
	// 日期是完整 YYYY-MM-DD。不要与正常发送账单合并维护。
	headers := []string{"交易日期", "记账日期", "交易说明", "交易币种/金额", "入账币种/金额"}
	return Adapter{
		Key:     "bocom_credit_reissue",
		Name:    "交通银行信用卡（补发账单）",
		Headers: headers,
		RowGuards: []RowGuard{
			{Column: 0, Format: RowGuardFormatYYYYDashMMDashDD},
			{Column: 1, Format: RowGuardFormatYYYYDashMMDashDD},
		},
	}
}
