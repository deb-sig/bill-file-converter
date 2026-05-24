# AGENTS.md

本文件只记录本项目的强约束。其他背景请读 `README.md` 和 `doc/`。

## 必须遵守

- CSV 只还原已支持账单类型的原始交易明细表；不要做跨银行统一 schema。
- `result.csv` 只能包含交易表头和交易行；不要写入标题、账单周期、卡号、
  metadata 或合成列。
- 找不到预期表格或表头时必须校验失败；不要输出猜测格式的 CSV。
- 修改账单类型时，必须以 MinerU `content_list` / fixture 的实际表格结构为准；
  不要只根据截图、legacy 代码或下游期望推断。
- `Adapter.Headers` 定义导出表头和 CSV 列顺序；`HeaderAliases` 只接受同一原始
  表头的 OCR/VLM 识别变体。
- `RowGuards` 只做结构性过滤，例如日期列或序号列；不要承载业务清洗、分类、
  去重或金额推断。
- 结构明显不同的账单格式必须拆成独立 `--type` key；不要在一个 adapter 中堆
  大量条件分支。
- 只有确认图片、水印或印章影响识别时才启用 `RemoveImages`。

## Fixture 和隐私

- 新增公共账单类型必须补 e2e fixture：`mineru_response.json`、`expected.csv`、
  `testdata/e2e/cases.json`，并更新 `doc/bill-types/` 文档。
- fixture 必须脱敏；不要提交真实账单、未脱敏 MinerU payload、secret、
  `config.yaml`、`output/` 或 `.private/`。
- 不要用简单去重修复行数异常。真实账单可能存在连续相同交易。

## 验收

- Go 代码改动后运行 `gofmt`。
- 大多数改动至少运行：

```bash
go test ./...
```

- 涉及 CLI 或 adapter 注册时，同时确认：

```bash
go run ./cmd/bill-file-converter list-types
```
