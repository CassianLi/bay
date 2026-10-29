# bay

本机项目港湾 CLI：登记、查询、巡检指定根目录下的工程，并用状态驱动物理位置（进入各根下的 `_pm/...`）。

设计说明：[docs/superpowers/specs/2026-10-29-bay-design.md](docs/superpowers/specs/2026-10-29-bay-design.md)

## 安装与配置（推荐）

```bash
# 在仓库根目录构建并进入引导安装
make local
./bay install
```

`bay install` 会：

1. 询问安装目录（默认 `~/bin` 或 `GOPATH/bin`），复制当前二进制为 `bay`
2. 检查该目录是否在 `PATH` 中，必要时打印加入方式
3. 若尚无配置（或你选择 / 使用 `--force-setup`），进入**交互式配置向导**
4. 可选立即执行第一次 `scan`

仅改配置、不重装二进制：

```bash
bay setup          # 配置不存在时
bay setup --force  # 覆盖已有配置
bay config show    # 查看当前配置
bay config path    # 打印配置文件路径
```

其它安装方式：

```bash
go install github.com/jinrunli/bay/cmd/bay@latest   # 需模块可拉取
go build -o bay ./cmd/bay && ./bay install --bin-dir ~/bin
```

## 默认路径

| 用途 | 路径 |
|------|------|
| 配置 | `~/.config/bay/config.yaml` |
| SQLite 登记库 | `~/.local/share/bay/ledger.db` |
| 默认管理根 | `~/CursorProjects`、`~/DeepseekProjects` |

可用全局 `--config` 覆盖配置路径；登记库始终使用上述默认 DB 路径。

## 命令

| 命令 | 说明 |
|------|------|
| `bay install` | 安装二进制 + 可选交互配置 / 首次扫描 |
| `bay setup` | 交互式写配置（roots / stale / pm_dir / ignore） |
| `bay config show` / `path` | 查看配置内容或路径 |
| `bay scan` | 扫描配置中的根目录，更新登记库 |
| `bay ls [--status S] [--unknown] [--root R] [keyword]` | 列出项目 |
| `bay show <name> [--root R]` | 显示单个项目详情 |
| `bay set <name> [name...] --status S` / `--summary` / `--note` | 单个或批量改状态、摘要、备注 |
| `bay doctor [--root R] [--md file]` | 巡检报告 |

## 状态与 `_pm` 映射

| 状态 | 物理位置 |
|------|----------|
| `idea` | `<root>/_pm/incubator/<name>` |
| `active` / `paused` / `kept` | `<root>/<name>` |
| `archived` | `<root>/_pm/archive/<name>` |
| `trash` | `<root>/_pm/trash/<name>`（不真删） |

旧根级 `_incubator` / `_archive` 仍可被 `scan` 识别；新的 `set` 搬动一律写入 `_pm/...`。不跨 root 搬迁。

## `set` 确认与试跑

- `--dry-run`：只打印搬迁计划，不搬盘、不写库
- `--yes`：跳过确认提示
- 默认：需要搬盘时提示 `proceed? [y/N]`；批量时合并确认一次

## 开发构建

```bash
make local     # 本机二进制 ./bay
make build     # linux/amd64 → dist/bay
make package   # dist 下 tar.gz
```
