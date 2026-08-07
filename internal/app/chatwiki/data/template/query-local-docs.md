---
name: query-local-docs
description: Query local reference documents stored in the skill's references directory. Use when the user asks about uploaded files, local documents, PDFs, DOCX files, Markdown files, product manuals, policy documents, README files, or any knowledge stored in the references folder.
---

# Query Local Docs

从 `references/` 目录下的本地文档中检索信息。

## 文档索引

```yaml
<index_yaml>
```

## 检索规则

根据索引中的 `description` 和 `keywords` 定位候选文件后，按以下流程执行：

1. **grep 搜索** — 对候选文件用关键词及近义词 grep，获取匹配行及上下文
2. **按需精读** — grep 片段不足时，读取匹配段落的更大上下文
3. **兜底全读** — 已 grep 至少 2 组关键词无结果且文件 ≤ 200 行时才允许

## 图片资源输出规则

框架会自动注入当前 skill 的实际目录。参考文档中的 `assets/...` 图片路径以该 Markdown 所在的 `references/`
目录为基准；回答需要输出图片时，必须拼接为真实文件路径，并在路径前增加 `/` 转换成 Web 路径，禁止直接使用原相对路径。

```markdown
原路径：![image](assets/示例文档-pdf-20260730120000.md/image.png)
输出路径：![image](/clawbot/skills_robot/exampleRobotKey/query-local-docs/references/assets/示例文档-pdf-20260730120000.md/image.png)
```

必须根据框架注入的实际 skill 目录动态推导，不能硬编码示例路径。

**禁止：**

- 未经 grep 直接 read 全文
- grep path 传入目录（必须为具体文件，如 `references/filename.ext`）
- 用 `ls`/`glob` 代替索引定位
- 多文件冲突时优先采信 `updated` 较新的文件
