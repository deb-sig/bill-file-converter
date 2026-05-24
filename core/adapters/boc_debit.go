package adapters

func bocDebitAdapter() Adapter {
	// 中行借记卡流水表格本身是 12 列。legacy 版本会额外追加借记卡号列；
	// 当前转换器以还原原始表格为原则，不把页眉卡号补进交易表。
	headers := []string{"记账日期", "记账时间", "币别", "金额", "余额", "交易名称", "渠道", "网点名称", "附言", "对方账户名", "对方卡号/账号", "对方开户行"}
	return Adapter{
		Key:     "boc_debit",
		Name:    "中国银行借记卡",
		Headers: headers,
		RowGuards: []RowGuard{
			{Column: 0, Format: RowGuardFormatYYYYDashMMDashDD},
		},
	}
}
