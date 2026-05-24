# 中国银行借记卡

> [!NOTE]
> 该账单类型尚未通过真实账单充分测试，需要志愿者帮助测试并反馈转换正常或异常的样本。

## 支持范围

| `--type` | 适用场景 |
| --- | --- |
| `boc_debit` | 中国银行借记卡交易明细 PDF |

## 下载账单 PDF

待补充。

## 转换命令

```bash
go run ./cmd/bill-file-converter convert ./boc-debit.pdf \
  --type boc_debit \
  --config config.yaml \
  --output output
```

## 输入注意事项

- 交易日期应为 `YYYY-MM-DD` 形态。
- 如果 PDF 包含多页交易明细，可以直接使用同一个 PDF；如果银行拆成多个 PDF，按顺序传入。

## 输出核对

CSV 表头应为：

```text
记账日期,记账时间,币别,金额,余额,交易名称,渠道,网点名称,附言,对方账户名,对方卡号/账号,对方开户行
```

`借记卡号` 不会作为额外列输出。标题、卡号等非表格文本可在 `result.json.metadata.raw_text` 中查看。
