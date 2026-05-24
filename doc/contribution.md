# 贡献指南

本文面向希望修改代码、补充账单类型或维护测试样例的贡献者。

## 项目结构

- `cmd/bill-file-converter`：CLI 入口。
- `cli`：命令解析、配置加载、输入路径展开、CLI 输出。
- `core`：转换主流程、MinerU client、表格清洗、校验、artifact 写入。
- `core/adapters`：内置账单类型 profile，定义 key、名称、表头、表头别名、行过滤规则和预处理开关。
- `core/logger`：阶段日志和排障 artifact。
- `testdata/e2e`：fake MinerU 响应、预期 CSV 和 e2e cases。
- `legacy`：旧的纯前端实现，仅作为历史参考。

CLI 使用 `adapters.BuiltinRegistry()`，其中只包含公开内置账单类型。私有账单类型应放在内置 registry 之外，并通过 `core.Options.AdapterRegistry` 传给 `core.Convert`。

## 新增账单类型流程

1. 准备真实 PDF 样本，优先使用银行 App、网银或邮件导出的原始 PDF，不要使用截图。
2. 使用现有 CLI 转换或 inspect，保留 `logger/content_list.json`。
3. 先看 MinerU 输出的表格块、表头形态、重复表头、页眉页脚和分组行，不要只根据截图或 legacy 代码推断。
4. 在 `core/adapters` 中新增账单类型 profile。
5. 补充 `testdata/e2e/<type>/mineru_response.json` 和 `expected.csv`。
6. 将新 case 加入 `testdata/e2e/cases.json`。
7. 更新 `doc/bill-types/README.md` 和对应账单类型文档。
8. 运行 `go test ./...`。

## Profile 设计原则

- CSV 应尽量还原已注册账单类型对应的原始账单表格布局，包括原始列集合和列顺序；不要为了兼容 legacy 输出、下游账务系统或跨银行统一格式而删列、补列或改列。
- 同一家银行、同一种卡，如果正常发送、补发、导出选项或文件来源导致表格结构明显不同，应优先拆成独立 `--type` key，而不是在同一个 profile 中堆叠大量条件分支。例如交通银行信用卡正常账单和补发账单分别使用 `bocom_credit_regular` 与 `bocom_credit_reissue`。
- `Headers` 表示最终导出的规范表头；`HeaderAliases` 只用于接受 MinerU/VLM 对同一张原始表头的不同识别形态，例如中英混排、空格差异或表头拆行后的英文行。
- `RowGuards` 只做结构性过滤，例如日期列、序号列；不要用它承载业务清洗、交易分类或金额推断。
- `RemoveImages` 只在图片浮层、水印或印章确实影响表格识别时启用。PDF 重写可能改变文档结构或文字映射。

私有 profile 最小示例：

```go
registry := adapters.NewRegistry(adapters.Adapter{
    Key:     "private_payroll",
    Name:    "Private Payroll",
    Headers: []string{"日期", "项目", "金额"},
    RowGuards: []adapters.RowGuard{
        {Column: 0, Format: adapters.RowGuardFormatYYYYDashMMDashDD},
    },
})
```

## 与 legacy 实现的有意差异

- legacy 可能在 CSV 中输出标题行；当前结果 CSV 只输出交易表头和数据行，标题、账单周期、卡号等非表格信息保留在 `result.json.metadata.raw_text` 中。
- `boc_debit` 不再额外追加 `借记卡号` 列，因为借记卡号来自页眉文本，不属于交易明细表原始列。
- `boc_credit_regular` / `boc_credit_reissue` 不再额外追加 `币种` 列，因为币种来自“人民币账户交易明细”“美元账户交易明细”等章节状态，不属于交易明细表原始列。后续如需记录币种，应通过状态机或 metadata 设计统一处理。
- `boc_credit_reissue` 保留原始 `MM/DD` 日期，不自动补全年份。后续如需补全年份，应通过 CLI 开关统一控制日期补全，而不是在 profile 中默认改写原始日期。

## 测试与验收

基础验收：

1. `go test ./...` 通过。
2. `go run ./cmd/bill-file-converter list-types` 能看到目标 key。
3. `go run ./cmd/bill-file-converter mineru test --config config.yaml` 成功。
4. `convert` 生成 `result/result.json` 和 `result/result.csv`。
5. 日志包含 task id，所有产物写入 `output/<task_id>/`。
6. `result.json.task_id` 与输出目录名一致。
7. `result.json.source.files` 记录源 PDF 路径和文件名，单个或多个 PDF 都使用同一结构。
8. `result.json.metadata.raw_text` 包含去重后的原始非表格文本。
9. `result.json.tables[].headers` 精确匹配账单类型允许的表头。
10. CSV 列顺序和行顺序与 MinerU 表格顺序一致，不包含 metadata 或 title 行。
11. 预期表格或表头不存在时，转换必须校验失败，不能生成猜测格式的 CSV。
12. 新增账单类型应至少补一组 fake MinerU e2e fixture；有真实参考 CSV 时，应记录真实 PDF 转换与参考 CSV 的行数、表头和字段差异。

开发新 fixture 时，如果样本中存在无法判断真伪的 VLM 膨胀重复行，不应把它作为唯一验收标准。应使用参考 CSV 或人工核对结果中的行数、金额合计和连续重复交易段。

## 后续 TODO

- 中国银行信用卡账单的币种信息来自“人民币账户交易明细”“美元账户交易明细”等章节，而不一定存在于交易表列中。后续如果要记录币种，可能需要在清洗阶段引入按内容块顺序推进的状态机，将章节状态附加到对应交易表或 metadata，而不是在 profile 中硬编码补列。
- 部分信用卡补制账单只在交易行中提供 `MM/DD` 日期，年份来自标题或账单周期。当前 profile 优先还原原始表格，不自动补全年份；后续如果要对齐 legacy 的年份补全逻辑，应考虑增加 CLI 开关，统一控制是否做日期补全，而不是在单个 profile 中默认改写原始日期。
