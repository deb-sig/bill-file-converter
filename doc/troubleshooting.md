# 排障指南

本文覆盖转换过程中最常见的问题。排障时请优先保留完整的 `output/<task_id>/` 目录。

## MinerU 连接失败或超时

先检查配置：

```bash
go run ./cmd/bill-file-converter mineru test --config config.yaml
```

常见原因：

- `mineru.base_url` 仍是默认示例 `http://127.0.0.1:<port>`。
- MinerU 服务未启动，或端口不是配置中的端口。
- 使用内网或三方兼容服务时，需要在 `mineru.headers` 中配置鉴权 header。
- 账单 PDF 较大或服务繁忙，`mineru.timeout` 太短。

客户端固定调用 `{base_url}/file_parse`。不要在 `base_url` 中重复追加 `/file_parse`。

## 缺少 Ghostscript

部分账单类型会先移除 PDF 中影响识别的图片浮层或水印，这需要系统中存在 `gs` 可执行文件。

如果日志中出现 Ghostscript 相关错误：

- macOS 可通过 Homebrew 安装：`brew install ghostscript`。
- 安装后确认 `gs` 在当前 shell 的 `PATH` 中：`which gs`。
- 如果账单类型没有启用图片移除预处理，则不需要 Ghostscript。

当前已知会使用该预处理的账单类型包括 `abc_debit` 和 `cmb_debit`。

## 未匹配表头或校验失败

如果转换提示 validation failed，通常表示 MinerU 输出中没有找到当前 `--type` 期望的交易明细表。

检查顺序：

1. 确认 `--type` 是否选对。正常账单和补发/补制账单可能是不同 key。
2. 打开 `logger/content_list.json`，查看是否存在 `type=table` 且带 `table_body` 的块。
3. 查看表头是否被 MinerU 识别成了不同形态，例如中英文混排、印章文字粘连、拆行丢字。
4. 确认导出的 PDF 是账单明细或流水明细，不是截图、回单、总览页或加密不可解析文件。

排障时不要直接修改 CSV 作为修复依据。应先确认 MinerU 原始表格输出是否可用。

## CSV 行数异常

当前方案依赖 MinerU/VLM 对 PDF 表格的识别结果。若原账单中存在连续多行完全相同或高度相似的交易，VLM 可能把这些行识别成错误数量，例如漏行或膨胀出过多重复行。

不要用简单去重修复这类问题，因为银行账单中连续重复交易可能是真实发生的交易，去重会误删真实数据。

建议核对：

- CSV 表头是否符合对应账单类型文档。
- 行数是否与银行导出页面或参考 CSV 一致。
- 金额合计是否合理。
- 连续重复交易段是否与原 PDF 一致。
- `result.json.tables[].source_pages` 是否覆盖了预期页码。

## 多 PDF 顺序异常

多个 PDF 通过命令行参数传入时，按参数顺序解析并合并。目录输入时，目录只展开第一层 `.pdf` 文件，并按文件名自然序排序。

建议文件命名：

```text
1.pdf
2.pdf
3.pdf
```

或：

```text
statement-001.pdf
statement-002.pdf
statement-003.pdf
```

避免混用不可排序的文件名，例如 `final.pdf`、`new.pdf`、`old.pdf`。

## 提交问题时需要提供的信息

如果需要报告问题，请尽量提供：

- 执行的完整命令。
- 使用的 `--type` key。
- `result/result.json`。
- `logger/content_list.json`。
- `logger/mineru_request.json` 和 `logger/mineru_response.json`。
- `logger/failure.json`，如果存在。

如账单包含敏感信息，请先脱敏。排障最关键的是 MinerU 输出的表格结构、表头形态、页码和错误信息。
