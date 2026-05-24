# 账单类型

账单类型 key 通过 CLI 的 `--type` 参数传入。请先按银行和卡类型选择文档，再根据文档中的说明选择具体 key。

## 支持列表

| 文档 | 支持的 `--type` |
| --- | --- |
| [中国农业银行借记卡](abc_debit.md) | `abc_debit` |
| [中国银行借记卡](boc_debit.md) | `boc_debit` |
| [中国银行信用卡](boc_credit.md) | `boc_credit_regular`, `boc_credit_reissue` |
| [交通银行借记卡](bocom_debit.md) | `bocom_debit` |
| [交通银行信用卡](bocom_credit.md) | `bocom_credit_regular`, `bocom_credit_reissue` |
| [招商银行借记卡](cmb_debit.md) | `cmb_debit` |
| [招商银行信用卡](cmb_credit.md) | `cmb_credit` |
| [众邦银行借记卡](zbank_debit.md) | `zbank_debit` |
| [中关村银行借记卡](zgc_debit.md) | `zgc_debit` |

查看当前二进制实际注册的账单类型：

```bash
go run ./cmd/bill-file-converter list-types
```

## 通用准备原则

- 优先使用银行 App、网银或邮箱提供的原始 PDF。
- 不要使用截图、拍照图片或手工重排后的 PDF。
- 同一期账单如果导出为多个 PDF，可以按页码顺序传入多个文件，或放入同一目录并使用可自然排序的文件名。
- 转换后请核对表头、行数、金额和连续重复交易段。
