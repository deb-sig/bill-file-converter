package adapters

func abcDebitAdapter() Adapter {
	// 农行 PDF 导出的日期和时间都是紧凑格式，且部分行交易时间为空，
	// 因此只用交易日期列判断是否为真实交易行。
	headers := []string{"交易日期", "交易时间", "交易摘要", "交易金额", "本次余额", "对手信息", "日志号", "交易渠道", "交易附言"}
	return Adapter{
		Key:     "abc_debit",
		Name:    "中国农业银行借记卡",
		Headers: headers,
		RowGuards: []RowGuard{
			// 交易日期提取为 YYYYMMDD，而不是 YYYY-MM-DD。
			{Column: 0, Format: RowGuardFormatYYYYMMDD},
		},
		// 农行账单可能包含影响表格提取的图片浮层或印章，解析前先去除图片。
		RemoveImages: true,
	}
}
