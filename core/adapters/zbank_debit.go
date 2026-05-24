package adapters

func zbankDebitAdapter() Adapter {
	// 众邦导出是一页一个 PDF。VLM 会把交易时间提取成紧凑时间戳，
	// 这里保持原始提取结果，不额外做格式化清洗。
	headers := []string{"交易时间", "币种", "交易金额", "账户余额", "对方姓名", "对方账号", "摘要"}
	// 红色电子回单章与“对方账号”表头重叠，VLM 经常会把印章文字
	// 合并进该表头单元格。
	headersWithStamp := []string{"交易时间", "币种", "交易金额", "账户余额", "对方姓名", "对方账号电子回单专用章", "摘要"}
	return Adapter{
		Key:           "zbank_debit",
		Name:          "众邦银行借记卡",
		Headers:       headers,
		HeaderAliases: [][]string{headersWithStamp},
		RowGuards: []RowGuard{
			// 示例提取值：2024123113:40:02。
			{Column: 0, Format: RowGuardFormatYYYYMMDDHHMMSS},
		},
	}
}
