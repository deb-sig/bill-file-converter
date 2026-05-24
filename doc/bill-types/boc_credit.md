# 中国银行信用卡

> [!NOTE]
> 该账单类型尚未通过真实账单充分测试，需要志愿者帮助测试并反馈转换正常或异常的样本。

## 支持范围

| `--type` | 适用场景 | 日期形态 |
| --- | --- | --- |
| `boc_credit_regular` | 正常信用卡账单 PDF | `YYYY-MM-DD` |
| `boc_credit_reissue` | 补制信用卡账单 PDF | `MM/DD` |

两者输出相同的交易明细列，但表格日期形态不同，因此需要选择正确 key。

## 下载账单 PDF

待补充。

## 转换命令

正常账单：

```bash
go run ./cmd/bill-file-converter convert ./boc-credit.pdf \
  --type boc_credit_regular \
  --config config.yaml \
  --output output
```

补制账单：

```bash
go run ./cmd/bill-file-converter convert ./boc-credit-reissue.pdf \
  --type boc_credit_reissue \
  --config config.yaml \
  --output output
```

## 输入注意事项

- 如果交易日期是完整日期，优先使用 `boc_credit_regular`。
- 如果交易日期只有 `MM/DD`，使用 `boc_credit_reissue`。
- 当前转换器不从标题或账单周期补全年份。
- 币种信息可能来自“人民币账户交易明细”“美元账户交易明细”等章节标题，不一定是交易表列；当前 CSV 不额外追加币种列。

## 输出核对

CSV 表头应为：

```text
交易日,银行记账日,卡号后四位,交易描述,存入,支出
```

标题、账单周期、币种章节文本会保留在 `result.json.metadata.raw_text` 中，交易 CSV 只包含交易表格。
