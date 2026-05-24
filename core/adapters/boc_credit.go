package adapters

func bocCreditRegularAdapter() Adapter {
	// 普通账单的明细表日期为 YYYY-MM-DD。legacy 版本会额外输出标题行和
	// 从人民币/外币章节推导出的币种列；当前 profile 先只还原交易明细表本身。
	headers := bocCreditHeaders()
	return Adapter{
		Key:     "boc_credit_regular",
		Name:    "中国银行信用卡（正常账单）",
		Headers: headers,
		RowGuards: []RowGuard{
			{Column: 0, Format: RowGuardFormatYYYYDashMMDashDD},
		},
	}
}

func bocCreditReissueAdapter() Adapter {
	// 补制账单的明细表日期为 MM/DD，年份只出现在账单标题中。为避免把标题
	// metadata 混入交易表，当前 profile 保留原始 MM/DD 日期。
	headers := bocCreditHeaders()
	return Adapter{
		Key:     "boc_credit_reissue",
		Name:    "中国银行信用卡（补制账单）",
		Headers: headers,
		RowGuards: []RowGuard{
			{Column: 0, Format: RowGuardFormatMMSlashDD},
		},
	}
}

func bocCreditHeaders() []string {
	return []string{"交易日", "银行记账日", "卡号后四位", "交易描述", "存入", "支出"}
}
