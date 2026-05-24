package adapters

func zgcDebitAdapter() Adapter {
	headers := []string{"交易流水号", "记账日期", "货币", "交易金额", "账户余额", "交易摘要", "对手信息"}
	// Do not enable RemoveImages for this profile. Ghostscript pdfwrite with
	// -dFILTERIMAGE keeps the table extractable, but rewrites this PDF's
	// STSong/UniGB text mapping in a way that makes MinerU metadata text garbled.
	return Adapter{
		Key:                          "zgc_debit",
		Name:                         "中关村银行借记卡",
		Headers:                      headers,
		BlankRowspanCarryoverColumns: []int{0},
		RowGuards: []RowGuard{
			{Column: 1, Format: RowGuardFormatYYYYDashMMDashDDHHMMSS},
		},
	}
}
