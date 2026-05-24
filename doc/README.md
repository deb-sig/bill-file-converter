# 文档

`bill-file-converter` 使用 MinerU local-compatible API 将银行账单 PDF 转换为 CSV。这里是项目的主文档入口。

## 使用者

- [使用指南](guide.md)：配置 MinerU、运行 CLI、理解输入输出和清洗规则。
- [账单类型](bill-types/README.md)：查看已支持的 `--type` key、下载账单 PDF 的方式和每类账单的注意事项。
- [排障指南](troubleshooting.md)：处理 MinerU 连接失败、Ghostscript 缺失、表头未匹配、行数异常等问题。

## 贡献者

- [贡献指南](contribution.md)：理解项目架构、新增账单类型、补充 e2e fixture、运行验收。

## 账单类型 key

CLI 使用 `--type` 指定账单类型，例如：

```bash
go run ./cmd/bill-file-converter convert ./statement.pdf \
  --type abc_debit \
  --config config.yaml
```

`--type` 是稳定的账单类型 key。它不是页面布局名，也不是银行名称的自由文本。当前支持列表以 `go run ./cmd/bill-file-converter list-types` 为准。
