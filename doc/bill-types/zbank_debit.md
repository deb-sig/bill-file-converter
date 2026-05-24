# 众邦银行借记卡

## 支持范围

| `--type` | 适用场景 |
| --- | --- |
| `zbank_debit` | 众邦银行借记卡交易明细 PDF |

## 下载账单 PDF

待补充。

## 转换命令

单文件：

```bash
go run ./cmd/bill-file-converter convert ./zbank-debit.pdf \
  --type zbank_debit \
  --config config.yaml \
  --output output
```

多文件：

```bash
go run ./cmd/bill-file-converter convert ./zbank-001.pdf ./zbank-002.pdf \
  --type zbank_debit \
  --config config.yaml \
  --output output
```

## 输入注意事项

- 多个 PDF 按命令行参数顺序解析并合并。
- 如果使用目录输入，请将文件命名为 `1.pdf`、`2.pdf` 或 `statement-001.pdf`、`statement-002.pdf`。
- VLM 可能把交易时间提取成紧凑时间戳，转换器保持原始提取结果。

## 输出核对

CSV 表头应为：

```text
交易时间,币种,交易金额,账户余额,对方姓名,对方账号,摘要
```

如果红色电子回单章与“对方账号”表头重叠，转换器会接受该表头变体，但输出仍使用标准表头。
