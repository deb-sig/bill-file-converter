# 中关村银行借记卡

## 支持范围

| `--type` | 适用场景 |
| --- | --- |
| `zgc_debit` | 中关村银行借记卡交易明细 PDF |

## 下载账单 PDF

待补充。

## 转换命令

```bash
go run ./cmd/bill-file-converter convert ./zgc-debit.pdf \
  --type zgc_debit \
  --config config.yaml \
  --output output
```

## 输入注意事项

- 不要对该类型启用额外的 Ghostscript 图片移除。该 PDF 的文字映射可能在重写后变成乱码，影响 MinerU metadata。
- 原始表格中部分交易流水号可能通过视觉 rowspan 跨行显示，转换会清理由 rowspan 带来的空值延续。

## 输出核对

CSV 表头应为：

```text
交易流水号,记账日期,货币,交易金额,账户余额,交易摘要,对手信息
```

重点核对交易流水号和记账日期是否与原 PDF 对齐。
