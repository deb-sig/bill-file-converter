# 交通银行信用卡

## 支持范围

| `--type` | 适用场景 | 表格特点 |
| --- | --- | --- |
| `bocom_credit_regular` | 正常发送的信用卡账单 PDF | 6 列，包含卡末四位，日期为 `MM/DD` |
| `bocom_credit_reissue` | 买单吧或邮箱补发账单 PDF | 5 列，不含卡末四位，日期为 `YYYY-MM-DD` |

正常账单和补发账单结构不同，不要混用 key。

## 下载账单 PDF

正常账单 PDF 下载方式待补充。

补发账单已知入口：

1. 打开交通银行信用卡买单吧 App，搜索“账单补发”。
2. 选择补发账单月份。
3. 确认邮箱地址并提交。
4. 从邮箱下载账单文件。
5. 如果收到的是邮件正文或 EML，需要先打印/导出为 PDF；可靠操作步骤待补充。

## 转换命令

正常账单：

```bash
go run ./cmd/bill-file-converter convert ./bocom-credit.pdf \
  --type bocom_credit_regular \
  --config config.yaml \
  --output output
```

补发账单：

```bash
go run ./cmd/bill-file-converter convert ./bocom-credit-reissue.pdf \
  --type bocom_credit_reissue \
  --config config.yaml \
  --output output
```

## 输入注意事项

- `bocom_credit_regular` 支持中英混合表头和英文表头行。
- `bocom_credit_reissue` 沿用补发账单的五列结构，不含卡末四位。
- 如果选错 key，通常会出现未匹配表头或校验失败。

## 输出核对

正常账单 CSV 表头应为：

```text
交易日期,记账日期,卡末四位,交易说明,交易币种/金额,入账币种/金额
```

补发账单 CSV 表头应为：

```text
交易日期,记账日期,交易说明,交易币种/金额,入账币种/金额
```

行数异常时参考 [排障指南](../troubleshooting.md)。
