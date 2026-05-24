package adapters

func cmbDebitAdapter() Adapter {
	// 只支持招行完整版导出。旧版简化导出会缺少“客户摘要”，
	// 下游 schema 容易产生歧义，因此直接拒绝。
	headers := []string{"记账日期", "货币", "交易金额", "联机余额", "交易摘要", "对手信息", "客户摘要"}
	// VLM 可能把双语表头提取成英文表头行；这里接受该形态，
	// 但输出仍统一为中文标准表头。
	englishHeaders := []string{"Date", "Currency", "Transaction Amount", "Balance", "Transaction Type", "Counter Party", "Customer Summary"}
	return Adapter{
		Key:           "cmb_debit",
		Name:          "招商银行借记卡",
		Headers:       headers,
		HeaderAliases: [][]string{englishHeaders},
		RowGuards: []RowGuard{
			// 招行借记卡真实交易行总是以记账日期开头。
			{Column: 0, Format: RowGuardFormatYYYYDashMMDashDD},
		},
		// 招行借记卡 PDF 本身是文本型，但水印是图片浮层，
		// 会影响 VLM 表格识别效果。
		RemoveImages: true,
	}
}
