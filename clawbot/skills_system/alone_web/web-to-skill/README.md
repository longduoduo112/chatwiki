# Web to Skill

Turn a public website or an explicit set of URLs into a searchable, reusable AI skill.

Web to Skill is designed for product documentation, help centers, public knowledge bases, and similar web content. It
discovers site navigation, captures dynamically rendered pages, creates local HTML snapshots and a JSONL retrieval
index, and packages everything as a standalone skill ZIP. The generated skill keeps the source pages as traceable
evidence while providing bounded local retrieval, so an AI agent can fetch only the pages it needs instead of loading an
entire site into context.

## Features

- **Single-entry discovery**: Provide one URL to discover pages from the site's navigation or documentation tree.
- **Explicit batch collection**: Provide multiple URLs to validate, normalize, and deduplicate only those URLs, without
  expanding the requested scope.
- **Dynamic rendering**: Capture JavaScript-rendered pages with Playwright and headless Chromium.
- **Site-specific extraction**: Built-in content selectors for ChatWiki Docs, Yuque, Feishu, OpenClaw Docs, Alibaba
  Cloud Help, KanCloud, and WeChat Official Account articles.
- **Resilient crawling**: Sequential processing, one retry for transient failures, redirect-target deduplication,
  consecutive-timeout stopping, and structured logs.
- **Content cleanup**: Remove scripts, advertisements, and other non-content elements, reject advertisement-only pages,
  extract metadata and keywords, and suppress high-frequency noise shared across pages.
- **Traceable indexing**: Every index record points to its source URL and saved HTML snapshot.
- **Current-scope updates**: Re-run normal URL preparation and rebuild the skill strictly from the resulting current URL
  list. A single entry URL rediscovers its directory, while an explicit URL batch keeps that supplied scope without
  directory discovery, so pages outside the current scope are not carried into the new package.
- **Validated ZIP reuse**: Treat the previous successful skill ZIP as a page cache. Reuse only current URLs whose index
  record and HTML snapshot are valid; crawl every cache miss or rejected snapshot normally.
- **Yuque error-page filtering**: On `www.yuque.com`, reject cached error pages, retry newly rendered error pages once,
  and omit persistent error pages from both HTML output and the index.
- **Stable update identity**: Preserve the existing skill name during updates while regenerating its description,
  summaries, topics, aliases, and other metadata from the rebuilt index.
- **Bounded metadata context**: Sample at most 60 page summaries proportionally across source sites before generating
  skill metadata.
- **Ready-to-use skill packages**: Generate `SKILL.md`, agent configuration, the web index, HTML snapshots, and
  retrieval helpers.

## How It Works

```mermaid
flowchart LR
    A[Public URLs] --> B[Prepare URL list]
    X[Previous skill ZIP, optional update cache] --> C
    B --> C[Reuse valid current pages and crawl misses]
    C --> D[HTML snapshots and JSONL index]
    D --> E[Validate crawl artifacts]
    E --> F[Build bounded metadata outline]
    F --> G[Create source metadata]
    G --> H[Skill ZIP]
```

Deterministic scripts handle URL preparation, crawling, crawl validation, metadata outlining, and packaging. Site-level
metadata such as the skill name, description, topic groups, and coverage notes must be produced by an AI agent from the
bounded outline. The build script validates and packages this metadata; it does not invent business information.

An update is a full current-directory rebuild, not an incremental merge. The freshly prepared URL list is the only scope
of the new skill. The previous ZIP can save network work, but pages found only in that ZIP are never copied into the new
output.

## Requirements

- Python 3.10+
- Network access to the target public website
- Playwright, Beautiful Soup, and jieba
- Playwright Chromium

Install the Python dependencies and browser:

```bash
python3 -m pip install playwright beautifulsoup4 jieba
python3 -m playwright install chromium
```

> The examples use `python3`. Replace it with `python` if that is the executable name in your environment.

## Quick Start

The examples below assume that the current directory is this `web-to-skill` skill directory and that all intermediate
artifacts are written to `./workspace`.

### 1. Prepare the URL List

Provide one entry URL to discover its documentation directory:

```bash
python3 scripts/prepare_urls.py \
  --out workspace/crawl/url-list.txt \
  "https://example.com/docs"
```

Alternatively, provide multiple explicit pages. Directory discovery is skipped in this mode:

```bash
python3 scripts/prepare_urls.py \
  --out workspace/crawl/url-list.txt \
  "https://example.com/docs/start" \
  "https://example.com/docs/configuration"
```

The result is a UTF-8 text file containing one normalized URL per line.

### 2. Crawl the Pages

```bash
python3 scripts/crawl_urls.py \
  --url-list workspace/crawl/url-list.txt \
  --out-dir workspace/crawl
```

This stage produces:

```text
workspace/crawl/
├── url-list.txt       # Final normalized URL list
├── index.jsonl        # URLs, titles, descriptions, keywords, and snapshot paths
├── crawl.log          # Discovery, progress, retries, failures, and stop reasons
└── html/              # Cleaned rendered HTML snapshots
```

Validate the crawl without exposing the complete log to model context:

```bash
python3 scripts/validate_crawl.py \
  --index workspace/crawl/index.jsonl
```

The helper resolves crawl-index HTML paths relative to `index.jsonl` and emits only the final crawl counts, a bounded
failure summary, a bounded redirect-duplicate summary, and a bounded Yuque error-page-skip summary.

Use debug mode to process at most the first five URLs while validating the workflow:

```bash
python3 scripts/crawl_urls.py \
  --url-list workspace/crawl/url-list.txt \
  --out-dir workspace/crawl-debug \
  --debug
```

The crawler writes the effective URL list into its output directory. In debug mode, this contains at most the first five
URLs, so the same validator can check the isolated debug artifacts:

```bash
python3 scripts/validate_crawl.py \
  --index workspace/crawl-debug/index.jsonl
```

### 3. Update an Existing Skill

First run the normal URL preparation step again so `workspace/crawl/url-list.txt` represents the current intended scope.
With one entry URL, this rediscovers the website's current directory. With an explicit URL batch, it validates,
normalizes, and deduplicates only the supplied pages without directory discovery. Then provide the most recent
successful skill ZIP as an optional cache:

```bash
python3 scripts/crawl_urls.py \
  --url-list workspace/crawl/url-list.txt \
  --out-dir workspace/crawl \
  --existing-skill workspace/existing-skill.zip \
  --expected-name example-docs
```

The crawler safely stages reusable data under `workspace/existing/` and creates a fresh `workspace/crawl/index.jsonl`
and `workspace/crawl/html/`. It copies only valid records that exactly match current normalized URLs. Missing,
unreadable, empty, Yuque error-page, or advertisement-only snapshots are rejected and fetched normally. Recognized
advertisement nodes are removed from otherwise reusable cached snapshots. If every current URL is reusable, the crawler
completes without launching Chromium.

Pages present only in the previous ZIP are excluded. If a newly fetched Yuque page is still an error page after one
retry, the page is skipped without writing HTML or an index record.

### 4. Create Skill Metadata

Generate a bounded metadata outline instead of loading the complete index:

```bash
python3 scripts/metadata_outline.py \
  --index workspace/crawl/index.jsonl
```

The helper returns at most 60 page summaries, allocated proportionally by source site and sampled evenly within each
site. Create `workspace/skill-metadata.json` using only that outline and the contract in
[`references/metadata.md`](references/metadata.md). Keep `coverage_notes` empty unless the sampled metadata explicitly
states a boundary; absence from a bounded outline is not evidence that a topic is unsupported.

For a new skill, generate all metadata normally. For an update, set the top-level `name` exactly to the existing skill
name supplied through `--expected-name`; regenerate descriptions, summaries, topics, aliases, and other metadata from
the current rebuilt index.

### 5. Build the Skill ZIP

```bash
python3 scripts/build_skill.py \
  --index workspace/crawl/index.jsonl \
  --metadata workspace/skill-metadata.json \
  --zip-out workspace/generate_skill/example-docs.zip
```

When updating, lock the package identity during the build:

```bash
python3 scripts/build_skill.py \
  --index workspace/crawl/index.jsonl \
  --metadata workspace/skill-metadata.json \
  --expected-name example-docs \
  --zip-out workspace/generate_skill/example-docs.zip
```

If `metadata.name` differs from `--expected-name`, correct that field and rerun only the build step; the crawl does not
need to run again.

The build script also requires `crawl.log` beside `index.jsonl`. It reads the last `crawl_urls run.done` event,
validates its counts against the index, and writes a deterministic reuse, success, failure, redirect-duplicate,
Yuque-error-page-skip, and timeout-skipped coverage note into the generated skill. A missing log or a log without
`run.done` fails the build.

The generated archive has the following structure:

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

The generated `search_index.py` returns a bounded set of candidate pages from the local index. `fetch_rendered_html.py`
refreshes a single page only when the saved snapshot is insufficient or current content is explicitly required.

## Site-Specific Behavior

| Site or content type | Strategy                                                                                                                                     |
|----------------------|----------------------------------------------------------------------------------------------------------------------------------------------|
| Generic public pages | Discover links from rendered navigation and fall back to the rendered page body                                                              |
| ChatWiki Docs        | Use the Docusaurus sitemap and retain the language selected by the entry URL                                                                 |
| KanCloud             | Read the full directory tree from `application/payload+json`; fail rather than silently continue when a known directory yields only one page |
| Yuque                | On `www.yuque.com`, reject cached error pages and retry then skip persistent newly rendered error pages                                      |
| Feishu               | Keep the longest stable body snapshot if the final rendered body becomes empty or shorter                                                    |
| Other adapted sites  | Use built-in body selectors and fall back to the rendered page body when necessary                                                           |

## Reliability and Scope

- Only public `http://` and `https://` pages are supported. The project does not handle authentication, CAPTCHAs, or
  access-control bypasses.
- URL preparation uses fixed 10-minute navigation and directory-collection limits. A failed navigation is retried once.
- URLs are crawled sequentially with a fixed 60-second page timeout. A timed-out page is retried once; crawling stops
  after four consecutive final timeouts.
- Prepared URLs that redirect to the same final page produce one index record. Redirect aliases are reported separately
  in crawl coverage instead of being counted as failures.
- During updates, the previous ZIP is only a validated cache for current URLs. Old-only pages are excluded, invalid
  cache entries fall back to a normal crawl, and an all-cache-hit update does not launch Chromium.
- Yuque error-page detection runs only for final URLs on `www.yuque.com`. It uses known error structures plus a guarded
  main-content text fallback, so ordinary articles containing words such as "sorry" are not skipped.
- Advertisement-only pages are retried once and omitted as crawl failures if they remain empty. Recognized advertisement
  nodes are removed from newly rendered and reusable cached snapshots.
- HTTP 429 and 5xx responses, browser network errors, timeouts, empty rendered bodies, Yuque error pages, and
  advertisement-only pages enter the retry path.
- URL preparation and crawling use fixed workflow policies. Concurrency, depth, link scope, timeout, and retry counts
  are intentionally not exposed as command-line controls.
- HTML snapshots represent the page state at crawl time. Re-run the crawler, or use the generated single-page refresh
  helper, when current content is required.
- Before crawling, make sure your use complies with the target site's terms of service, robots policies, and content
  permissions.

## Project Structure

```text
.
|-- README.md
|-- SKILL.md                    # Agent workflow
|-- agents/openai.yaml          # Skill presentation and default prompt
|-- references/metadata.md      # Model-authored metadata schema and limits
`-- scripts/
    |-- prepare_urls.py         # URL validation, normalization, and discovery
    |-- crawl_urls.py           # Validated cache reuse, crawling, Yuque filtering, indexing, and logs
    |-- validate_crawl.py       # Relative-path checks and bounded crawl summary
    |-- metadata_outline.py     # Bounded proportional metadata sampling
    |-- build_skill.py          # Metadata validation and ZIP packaging
    |-- search_index.py         # Bounded local retrieval for generated skills
    `-- fetch_rendered_html.py  # Playwright rendering and single-page refresh
```

## Design Principles

- **Traceable facts**: Retrieved information remains connected to a source URL and saved HTML snapshot.
- **Controlled scope**: Single-entry discovery and explicit URL batches use separate strategies to prevent accidental
  crawl expansion.
- **Fresh update scope**: The current URL list defines every rebuild; previous output is reusable input, never an
  authority for retaining deleted pages.
- **Bounded context**: Metadata generation uses a source-proportional outline, while the generated JSONL index supports
  search-first, read-later access instead of loading the entire site into an agent context.
- **Portable output**: Each final ZIP includes its index, snapshots, and runtime helpers, making it suitable for
  independent distribution and installation.
