# binary-parser

解析一种简单的分块二进制容器格式（文件头 `BCHK`），输出结构摘要或树状视图，并校验每条记录的 CRC32。

## 特性

- 解析文件头（magic / 版本 / 记录数）
- 逐条解析记录（类型 / ID / 载荷 / CRC32）
- 按类型分组、按 ID 索引
- 树状结构可视化与 CRC 校验状态

## 用法

```bash
# 文本摘要
binary-parser parse sample.bchk

# JSON 摘要
binary-parser parse sample.bchk --format json

# 树状结构 + CRC 状态
binary-parser parse sample.bchk --tree
```

输入文件缺失或格式非法时返回受控错误（退出码非 0），不会崩溃。

## 构建

```bash
go build ./...
go test ./...
```
