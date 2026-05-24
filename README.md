# bill-file-converter

[![License](https://img.shields.io/github/license/deb-sig/bill-file-converter)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/deb-sig/bill-file-converter)](go.mod)

使用 OCR/VLM 将银行账单 PDF 转换为 CSV 的命令行工具。

> [!NOTE]
> 旧版本基于 pdf.js 的纯前端实现位于 `legacy-pdfjs` 分支

## Features

- PDF 输入：支持单个 PDF、多个 PDF 或一层目录输入。
- OCR/VLM 解析：适合处理银行 PDF 中复杂表格、双语表头、水印和印章干扰。
- CSV 输出：还原已支持账单类型的原始交易明细表，不做跨银行交易归一化。
- 多银行账单：支持中国银行、交通银行、招商银行、农业银行、众邦银行、中关村银行等账单类型。
- 可扩展：通过 Go 代码新增账单类型处理规则，并用 e2e fixture 验证。

## Quick Start

运行或构建 CLI：

```bash
go run ./cmd/bill-file-converter list-types
go build ./cmd/bill-file-converter
```

生成配置文件并填写 OCR/VLM 服务 endpoint：

```bash
go run ./cmd/bill-file-converter config init -output config.yaml
```

检查解析服务：

```bash
go run ./cmd/bill-file-converter mineru test --config config.yaml
```

转换 PDF：

```bash
go run ./cmd/bill-file-converter convert ./statement.pdf \
  --type cmb_debit \
  --config config.yaml \
  --output output
```

生成结果位于 `output/<task_id>/result/result.csv`，排障文件位于 `output/<task_id>/logger/`。

## Supported Bill Types

完整列表见 [账单类型文档](doc/bill-types/README.md)。当前内置 key 包括：

```text
abc_debit
boc_debit
boc_credit_regular
boc_credit_reissue
bocom_debit
bocom_credit_regular
bocom_credit_reissue
cmb_debit
cmb_credit
zbank_debit
zgc_debit
```

运行以下命令可查看当前二进制实际注册的账单类型：

```bash
go run ./cmd/bill-file-converter list-types
```

## Documentation

- [文档首页](doc/README.md)
- [使用指南](doc/guide.md)
- [排障指南](doc/troubleshooting.md)
- [账单类型](doc/bill-types/README.md)
- [贡献指南](doc/contribution.md)

## Scope

- 当前只支持 PDF 输入。邮件账单请先在邮箱客户端或浏览器中打印/导出为 PDF。
- 程序只消费 OCR/VLM 解析结果中的表格块。
- CSV 只负责还原已支持账单类型的原始表格布局，不输出标题行，不补跨银行统一字段。
- metadata 仅保留去重后的原始非表格文本，写入 `result.metadata.raw_text`，用于调试。

更多边界和已知限制见 [使用指南](doc/guide.md) 与 [排障指南](doc/troubleshooting.md)。

## Contributing

新增账单类型、测试 fixture 和开发流程见 [贡献指南](doc/contribution.md)。

## Acknowledgements

当前解析服务适配 [MinerU](https://github.com/opendatalab/MinerU) local-compatible API。感谢 MinerU 项目提供的 PDF 解析能力。

## License

[Apache-2.0](LICENSE)
