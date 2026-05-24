package adapters

func bocomDebitAdapter() Adapter {
	headers := []string{"序号", "交易日期", "交易时间", "交易类型", "借贷", "交易金额", "余额", "对方账号", "对方户名", "交易地点", "摘要"}
	bilingualHeaders := []string{
		"Serial Num 序号",
		"Trans Date 交易日期",
		"Trans Time 交易时间",
		"Trading Type 交易类型",
		"Dc Flg 借贷",
		"Trans Amt 交易金额",
		"Balance 余额",
		"Payment Receipt Account 对方账号",
		"Payment Receipt Account Name 对方户名",
		"Trading Place 交易地点",
		"Abstract 摘要",
	}
	truncatedBilingualHeaders := []string{
		"Num 序号",
		"Trans Date 交易日期",
		"Trans Time 交易时间",
		"Trading Type 交易类型",
		"Dc Flg 借贷",
		"Trans Amt 交易金额",
		"Balance 余额",
		"Account 对方账号",
		"Account Name 对方户名",
		"Trading Place 交易地点",
		"Abstract 摘要",
	}
	return Adapter{
		Key:           "bocom_debit",
		Name:          "交通银行借记卡",
		Headers:       headers,
		HeaderAliases: [][]string{bilingualHeaders, truncatedBilingualHeaders},
		RowGuards: []RowGuard{
			// 交行明细每条交易都有递增序号；页眉、页脚和“打印完毕”不满足该结构。
			{Column: 0, Format: RowGuardFormatPositiveInteger},
		},
	}
}
