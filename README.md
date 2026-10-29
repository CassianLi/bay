# bay

本机项目港湾 CLI：登记、查询、巡检指定根目录下的工程，并用状态驱动物理位置（进入各根下的 `_pm/...`）。

设计说明：[docs/superpowers/specs/2026-10-29-bay-design.md](docs/superpowers/specs/2026-10-29-bay-design.md)

## 安装

```bash
# 安装到 GOPATH/bin（或 GOBIN）
go install github.com/jinrunli/bay/cmd/bay@latest

# 或在仓库根目录构建本地二进制
go build -o bay ./cmd/bay
```

## 默认路径

| 用途 | 路径 |
|------|------|
| 配置 | `~/.config/bay/config.yaml` |
| SQLite 登记库 | `~/.local/share/bay/ledger.db` |
| 默认管理根 | `~/CursorProjects`、`~/DeepseekProjects` |

首次运行若配置文件不存在，会写入默认配置（`stale_after_days: 90`，`pm_dir: _pm`）。可用全局 `--config` 覆盖配置路径；登记库始终使用上述默认 DB 路径。

## 命令

| 命令 | 说明 |
|------|------|
| `bay scan` | 扫描配置中的根目录，更新登记库 |
| `bay ls [--status S] [--unknown] [--root R] [keyword]` | 列出项目（默认按 `name` 升序） |
| `bay show <name> [--root R]` | 显示单个项目详情 |
| `bay set <name> --status S` / `--summary TXT` `[--root R]` | 改状态和/或用途摘要；状态变更可能搬目录 |
| `bay doctor [--root R] [--md file]` | 巡检报告（终端文本；可选 Markdown 落盘） |

## 状态与 `_pm` 映射

| 状态 | 物理位置 |
|------|----------|
| `idea` | `<root>/_pm/incubator/<name>` |
| `active` / `paused` / `kept` | `<root>/<name>` |
| `archived` | `<root>/_pm/archive/<name>` |
| `trash` | `<root>/_pm/trash/<name>`（不真删） |

旧根级 `_incubator` / `_archive` 仍可被 `scan` 识别；新的 `set` 搬动一律写入 `_pm/...`。不跨 root 搬迁。

## `set` 确认与试跑

- `--dry-run`：只打印搬迁计划（`from → to`），不搬盘、不写库
- `--yes`：跳过确认提示，直接执行搬迁
- 默认：需要搬盘时提示 `proceed? [y/N]`，仅 `y` / `yes` 继续；目标路径已存在则中止
