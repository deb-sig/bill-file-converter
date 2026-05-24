package adapters

func cmbCreditAdapter() Adapter {
	// 招行信用卡账单使用固定六列表格；还款/退款/消费这类分组行
	// 不在 adapter 中建模，而是通过行判定自然过滤掉。
	headers := []string{"交易日", "记账日", "交易摘要", "人民币金额", "卡号末四位", "交易地金额"}
	// VLM 会把中英文表头合并进同一个单元格。
	bilingualHeaders := []string{
		"交易日 SOLD",
		"记账日 POSTED",
		"交易摘要 DESCRIPTION",
		"人民币金额 RMB AMOUNT",
		"卡号末四位 CARD NO(Last 4digits)",
		"交易地金额 Original Tran Amount",
	}
	return Adapter{
		Key:           "cmb_credit",
		Name:          "招商银行信用卡",
		Headers:       headers,
		HeaderAliases: [][]string{bilingualHeaders},
		RowGuards: []RowGuard{
			// 还款行的交易日可能为空，但记账日稳定存在。
			{Column: 1, Format: RowGuardFormatMMSlashDD},
		},
	}
}
