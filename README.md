# bay

[![CI](https://github.com/CassianLi/bay/actions/workflows/ci.yml/badge.svg)](https://github.com/CassianLi/bay/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**bay** — 本机「项目港湾」CLI。给 vibe coding 时代散落的小工具仓做登记、查询与巡检，并用状态驱动目录位置（`_pm/incubator|archive|trash`），减少「做过但忘了」的资产流失。

> Local project harbor: inventory directories under configured roots, track purpose/status, and move projects by lifecycle — without deleting anything automatically.

## 功能概览

- **scan**：扫描配置的工程根，写入 SQLite 登记库（轻量信号：README 摘录、git 时间、栈标签）
- **ls / show**：按状态、关键词查看；列表对齐显示简要说明
- **set**：单个或**批量**改状态 / 用途摘要 / **备注**；改状态可联动挪目录（`--dry-run` / `--yes`）
- **doctor**：未知用途、久未动、trash、失踪项报告
- **install / setup**：安装二进制 + **交互式配置向导**

## 快速开始

```bash
git clone git@github.com:CassianLi/bay.git
cd bay
make local
./bay install
```

`bay install` 会复制二进制到 `~/bin`（或 `GOPATH/bin`）、检查 `PATH`，并引导配置 `roots` / `stale_after_days` / `pm_dir`，可选首次 `scan`。

仅配置：

```bash
bay setup          # 新建配置
bay setup --force  # 覆盖已有配置
bay config show
```

也可用：

```bash
go install github.com/CassianLi/bay/cmd/bay@latest
bay setup
bay scan
```

## 默认路径

| 用途 | 路径 |
|------|------|
| 配置 | `~/.config/bay/config.yaml` |
| 登记库 | `~/.local/share/bay/ledger.db` |
| 默认 roots | `~/CursorProjects`、`~/DeepseekProjects` |

全局 `--config` 可覆盖配置文件路径；登记库使用上述默认 DB 路径。

## 命令

| 命令 | 说明 |
|------|------|
| `bay install` | 安装二进制 + 可选交互配置 / 首次扫描 |
| `bay setup` | 交互式写配置 |
| `bay config show` / `path` | 查看配置或路径 |
| `bay scan` | 扫描 roots，更新登记库 |
| `bay ls [--status S] [--unknown] [--root R] [keyword]` | 列表（`NAME STATUS SUMMARY PATH`） |
| `bay show <name> [--root R]` | 单条详情 |
| `bay set <name> [name...] --status S` / `--summary` / `--note` | 单个或批量更新 |
| `bay doctor [--root R] [--md file]` | 巡检报告 |

状态取值：`idea|active|paused|kept|archived|trash`（见 `bay set -h`）。

## 状态与目录

| 状态 | 物理位置 |
|------|----------|
| `idea` | `<root>/_pm/incubator/<name>` |
| `active` / `paused` / `kept` | `<root>/<name>` |
| `archived` | `<root>/_pm/archive/<name>` |
| `trash` | `<root>/_pm/trash/<name>`（隔离，不真删） |

旧根级 `_incubator` / `_archive` 仍可被 `scan` 识别；新的 `set` 搬动写入 `_pm/...`。不跨 root 搬迁。

## 开发

```bash
make local     # 本机 ./bay
make build     # linux/amd64 → dist/bay
make package   # dist 下 tar.gz
go test ./...
```

更多约定见 [CONTRIBUTING.md](CONTRIBUTING.md)。设计文档索引：[docs/README.md](docs/README.md)。

## 安全与隐私

bay 只读写本机配置与登记库，不会上传数据。报告问题见 [SECURITY.md](SECURITY.md)。

## License

[MIT](LICENSE) © Cassian Li
