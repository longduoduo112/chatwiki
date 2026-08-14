![横幅](images/banner.png)

Web to Skill 专为产品文档、帮助中心、公共知识库及类似 Web 内容而设计。它能够发现网站导航、捕获动态渲染页面、创建本地 HTML 快照和
JSONL 检索索引，并将所有内容打包为独立的 Skill ZIP。生成的技能保留源页面作为可追溯证据，同时提供有界的本地检索，因此 AI
Agent 只需获取所需页面，无需加载整个网站。

> 官网知识和帮助文档沉睡在 HTML 中；Web to Skill 让网站文档成为客服 Agent 可随时调用的技能。

---

## 为什么需要 Web to Skill

企业的产品、帮助、活动和政策散落在不同页面，客户却需要统一、准确的答案。传统方式很难直接把网页转变为 Agent 的对话能力：

- **官网内容分散各处**：产品、帮助中心、博客和活动散布在不同栏目，人工难以提供统一且一致的回答。
- **维护成本高且容易过期**：依靠人工整理和同步，更新慢、易遗漏，网站改版后知识很快失效。
- **Agent 无法直接使用官网**：原始网页不能直接被 Agent 高效理解和调用，客服只能凭经验手工回答。

未经结构化处理时，Agent 每次回答都可能把大量相关网页的 HTML 放入上下文，令牌消耗大、响应慢、成本高。生成 Skill
后，可按页面和主题精准检索，只调取需要的内容。

---

![流程](images/process.png)

## 核心特性

- **单入口发现**：提供一个 URL，即可从网站导航或文档树发现同一范围内的页面。
- **显式批量收集**：提供多个 URL 时，只校验、规范化和去重这些 URL，不会扩大请求范围。
- **动态渲染**：通过 Playwright 和无头 Chromium 捕获 JavaScript 渲染后的页面。
- **特定站点提取**：内置 ChatWiki 文档、语雀、飞书、OpenClaw 文档、阿里云帮助、看云和微信公众号文章的内容选择器。
- **弹性抓取**：顺序处理、瞬态故障一次重试、重定向目标去重、连续超时停止，并记录结构化日志。
- **内容清理**：移除脚本、广告和其他非内容元素；过滤纯广告页面，提取元数据和关键词，并抑制跨页面高频噪声。
- **可追溯索引**：每条索引记录都关联源 URL 和保存的 HTML 快照。
- **当前范围更新**：每次更新都会从当前 URL 列表重建技能；旧 ZIP 只作为当前页面的已验证缓存，不会把已不在当前范围内的页面带入新包。
- **语雀错误页过滤**：对 `www.yuque.com` 的错误页进行重试，持续错误页不会写入 HTML 快照或索引。
- **稳定的更新标识**：更新时保留既有技能名称，同时根据当前重建后的索引重新生成其他元数据。
- **有界元数据上下文**：生成技能元数据前，按源站点比例最多抽取 60 条页面摘要。
- **即用型技能包**：生成 `SKILL.md`、Agent 配置、Web 索引、HTML 快照和检索辅助程序。

---

## 适用场景

- **售前产品咨询**：基于官网产品与方案页，精准回答参数、价格和适用场景，引导客户完成购买决策。
- **帮助中心自助答疑**：抓取 FAQ 与文档，让客户自助查找问题和解决办法，提升满意度、降低人工成本。
- **营销活动解答**：同步官网活动与品牌动态，准确解答优惠规则、参与方式和截止时间。
- **渠道合作与下载引导**：汇总合作政策和下载资源，自动引导合作伙伴和用户找到正确入口与资料。

---

## 工作原理

确定性脚本负责 URL 准备、抓取、抓取验证、元数据大纲和打包。站点级元数据（如技能名称、描述、主题组和覆盖范围说明）必须由 AI Agent
根据有界大纲生成；构建脚本只验证和打包这些元数据，不会虚构业务信息。

对于单个入口 URL，工具会发现该页面所属目录或导航树中的页面；对于多个明确 URL，工具只处理提供的页面。随后将每个成功页面的清洗后
HTML 与可追溯 JSONL 索引打包为技能。使用时，Agent 先检索相关页面，再按需读取其快照内容，从官网权威材料中作答。

更新是一次完整的当前范围重建，而非增量合并。最新 URL 列表是新技能的唯一范围；旧 ZIP 可以节省网络抓取，但只存在于旧包中的页面绝不会复制到新输出中。

---

## 环境要求与安装

**要求**

- Python 3.10+
- 可访问目标公共网站的网络
- `playwright`、`beautifulsoup4` 和 `jieba`
- Playwright 的 Chromium 浏览器

**安装 Python 依赖和浏览器：**

```bash
python3 -m pip install playwright beautifulsoup4 jieba
python3 -m playwright install chromium
```

> 示例使用 `python3`；若你的环境中可执行文件名为 `python`，请相应替换。

---

## 快速开始

以下示例假设当前目录是此 `web-to-skill` 技能目录，所有中间工件均写入 `./workspace`。

### 1. 准备 URL 列表

提供一个入口 URL 以发现其文档目录：

```bash
python3 scripts/prepare_urls.py \
  --out workspace/crawl/url-list.txt \
  "https://example.com/docs"
```

或者提供多个明确页面；此模式跳过目录发现：

```bash
python3 scripts/prepare_urls.py \
  --out workspace/crawl/url-list.txt \
  "https://example.com/docs/start" \
  "https://example.com/docs/configuration"
```

结果是 UTF-8 文本文件，每行包含一个规范化 URL。目录导航和目录收集使用固定的 10 分钟限制；导航失败会重试一次。不要手动创建或编辑
URL 列表。

### 2. 抓取页面

```bash
python3 scripts/crawl_urls.py \
  --url-list workspace/crawl/url-list.txt \
  --out-dir workspace/crawl
```

此阶段产生：

```text
workspace/crawl/
├── url-list.txt       # 最终规范化 URL 列表
├── index.jsonl        # URL、标题、描述、关键词和快照路径
├── crawl.log          # 发现、进度、重试、失败和停止原因
└── html/              # 清洗后的渲染 HTML 快照
```

可在不将完整日志暴露给模型上下文的情况下验证抓取：

```bash
python3 scripts/validate_crawl.py \
  --index workspace/crawl/index.jsonl
```

该辅助程序会相对于 `index.jsonl` 解析 HTML 路径，并仅返回最终抓取计数、有界失败摘要、有界重定向重复摘要和有界语雀错误页跳过摘要。

调试模式最多处理前五个 URL，适合验证工作流：

```bash
python3 scripts/crawl_urls.py \
  --url-list workspace/crawl/url-list.txt \
  --out-dir workspace/crawl-debug \
  --debug
```

调试输出目录会写入实际处理的 URL 列表，因此同样可以验证隔离的调试工件：

```bash
python3 scripts/validate_crawl.py \
  --index workspace/crawl-debug/index.jsonl
```

更新已有技能时，先按当前意图重新执行上一步，使 `workspace/crawl/url-list.txt` 表示当前范围；单入口会重新发现当前目录，显式
URL 批次只保留当前提供的页面。然后将最近一次成功生成的技能 ZIP 作为可选缓存传入：

```bash
python3 scripts/crawl_urls.py \
  --url-list workspace/crawl/url-list.txt \
  --out-dir workspace/crawl \
  --existing-skill workspace/existing-skill.zip \
  --expected-name example-docs
```

爬虫会在 `workspace/existing/` 下安全暂存可复用数据，并为本次运行创建新的 `workspace/crawl/index.jsonl` 与
`workspace/crawl/html/`。只有与当前规范化 URL 完全匹配且 HTML
可读取、非语雀错误页、非纯广告页的记录才会复用；其他页面正常抓取。若当前页面全部命中可复用缓存，仍会生成完整日志和新索引，但不会启动
Chromium。不要手动编辑 URL 列表、索引、日志或 HTML 快照。

### 3. 创建技能元数据

生成有界元数据大纲，而不是加载完整索引：

```bash
python3 scripts/metadata_outline.py \
  --index workspace/crawl/index.jsonl
```

该辅助程序最多返回 60 条页面摘要，按源站点成功页面数比例分配，并在每个站点内均匀抽样。仅使用该大纲和 [
`references/metadata.md`](references/metadata.md) 中的约定创建 `workspace/skill-metadata.json`。除非抽样元数据明确声明边界，否则将
`coverage_notes` 留空；有界大纲中没有某主题并不表示该主题不受支持。

新技能按常规生成全部元数据。更新时，顶层 `name` 必须与 `--expected-name` 中提供的既有技能名称完全一致；描述、摘要、主题、别名和其他元数据则应根据当前重建后的索引重新生成。

### 4. 构建技能 ZIP

```bash
python3 scripts/build_skill.py \
  --index workspace/crawl/index.jsonl \
  --metadata workspace/skill-metadata.json \
  --zip-out workspace/generate_skill/example-docs.zip
```

更新时锁定包标识：

```bash
python3 scripts/build_skill.py \
  --index workspace/crawl/index.jsonl \
  --metadata workspace/skill-metadata.json \
  --expected-name example-docs \
  --zip-out workspace/generate_skill/example-docs.zip
```

若 `metadata.name` 与 `--expected-name` 不一致，只更正该字段并重新构建，无需重新抓取。构建脚本还要求 `index.jsonl` 旁存在
`crawl.log`：它会读取最后一个 `crawl_urls` 的 `run.done`
事件，以索引验证其计数，并把确定性的复用、成功、失败、重定向重复、语雀错误页跳过和超时跳过覆盖说明写入生成的技能。日志缺失或不完整都会导致构建失败。

---

## 输出结构

生成的归档文件具有以下结构：

```text
example-docs/
├── SKILL.md
├── agents/
│   └── openai.yaml
├── references/
│   ├── web-index.jsonl
│   └── html/
│       └── *.html
└── scripts/
    ├── search_index.py
    └── fetch_rendered_html.py
```

生成的 `search_index.py` 会从本地索引返回一组有界的候选页面。`fetch_rendered_html.py` 仅在保存的快照不足，或明确需要当前内容时刷新单个页面。

---

## 项目结构

```text
.
|-- README.md
|-- SKILL.md                    # Agent 工作流
|-- agents/openai.yaml          # 技能展示与默认提示词
|-- references/metadata.md      # 模型编写的元数据结构与限制
`-- scripts/
    |-- prepare_urls.py         # URL 校验、规范化和发现
    |-- crawl_urls.py           # 已验证缓存复用、抓取、语雀过滤、索引和日志
    |-- validate_crawl.py       # 相对路径检查与有界抓取摘要
    |-- metadata_outline.py     # 有界的按比例元数据抽样
    |-- build_skill.py          # 元数据校验与 ZIP 打包
    |-- search_index.py         # 生成技能的有界本地检索
    `-- fetch_rendered_html.py  # Playwright 渲染与单页刷新
```

## 站点特定行为

| 网站或内容类型 | 策略                                                                                       |
|----------------|--------------------------------------------------------------------------------------------|
| 通用公共页面   | 从渲染后的导航中发现链接，必要时回退到渲染后的页面正文。                                   |
| ChatWiki 文档  | 使用 Docusaurus 网站地图，并保留入口 URL 所选择的语言。                                    |
| 看云           | 从 `application/payload+json` 读取完整目录树；已知目录只返回一页时直接失败，不会静默继续。 |
| 语雀           | 对 `www.yuque.com` 的缓存错误页予以拒绝；新渲染错误页重试一次，仍失败则跳过。              |
| 飞书           | 若最终渲染正文变为空或变短，保留最长的稳定正文快照。                                       |
| 其他已适配网站 | 使用内置正文选择器，必要时回退到渲染后的页面正文。                                         |

---

## 设计原则

- **可追溯的事实**：检索结果始终关联源 URL 和保存的 HTML 快照。
- **受控范围**：单入口发现和显式 URL 批处理采用不同策略，避免意外扩大抓取范围。
- **当前范围更新**：每次重建由当前 URL 列表定义范围；旧输出只是可复用输入，不能作为保留已删除页面的依据。
- **有界上下文**：元数据生成使用与源站点成比例的大纲；JSONL 索引支持先搜索、后读取，而不是把整个网站装入 Agent 上下文。
- **便携式输出**：每个最终 ZIP 都包含索引、快照和运行时辅助程序，可独立分发和安装。

---

## 可靠性与范围

- 仅支持公共 `http://` 和 `https://` 页面；不处理身份验证、验证码或访问控制绕过。
- URL 准备的导航与目录收集使用固定的 10 分钟限制；导航失败会重试一次。
- URL 按顺序抓取，单页固定超时 60 秒；超时页面重试一次，连续四次最终超时后停止抓取。
- 指向同一最终页面的已准备 URL 只会生成一条索引记录；重定向别名会在抓取覆盖率中单独报告，不计为失败。
- 更新时，旧 ZIP 仅是当前 URL 的已验证缓存；旧包独有页面会被排除，无效缓存会回退到正常抓取，全部命中缓存时不会启动 Chromium。
- 语雀错误页检测仅针对最终 URL 为 `www.yuque.com` 的页面，使用已知错误结构和受保护的正文文本回退，因此不会因普通文章出现“sorry”等词而被跳过。
- 纯广告页面会重试一次；若仍为空则记为抓取失败并不写入快照和索引。新页面与可复用缓存中的已知广告节点都会被移除。
- HTTP 429、5xx 响应、浏览器网络错误、超时、空渲染正文、语雀错误页和纯广告页都会进入重试路径。
- URL 准备和抓取使用固定工作流策略；并发、深度、链接范围、超时和重试次数不会作为命令行参数开放。
- HTML 快照反映抓取时的页面状态；需要最新内容时，请重新运行爬虫，或使用生成的单页刷新辅助程序。
- 抓取前请确保使用方式符合目标网站的 **服务条款、robots 政策和内容权限**。

---
