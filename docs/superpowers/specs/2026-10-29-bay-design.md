# bay — 本机项目港湾设计

日期：2026-10-29  
状态：实现完成  
形态：Go CLI + SQLite 主库 + YAML 配置

## 1. 产品定位与边界

### 是什么

`bay` 是本机「项目港湾」命令行工具：登记、查询、巡检你在指定根目录下的工程，并用状态驱动物理位置（进入各根下的 `_pm/...`），减少 vibe coding 带来的「做过但忘了」资产流失与垃圾堆积。

### 要解决的问题

- 打开陌生目录时，不知道它当初干什么、现在什么状态
- 分不清真实资产、搁置项目、可归档 / 可删项目
- 需要按关键词或状态反查历史工具
- 需要定期拿到「该归档 / 该进 trash / 用途未知」清单

### 非目标（第一版不做）

- 不自动删除磁盘内容（`trash` 只隔离到 `_pm/trash`）
- 不做重型代码理解或 LLM 推断用途
- 不监视文件系统、不安装 git hook
- 不管理配置根之外的目录
- 不做 GUI、多用户、远程同步
- 不做跨 root 搬迁、不做项目内双写元数据（`.bay.yaml`）

### 真相分层

| 层 | 职责 |
|----|------|
| SQLite 登记库 | 状态、用途、purpose、路径、扫描信号、时间戳 → **查询与巡检的真相** |
| 磁盘位置 | 服从状态映射；`set` 可搬目录并回写路径 |
| YAML 配置 | 根目录列表、忽略规则、状态↔目录映射、陈旧阈值 |

## 2. 目录约定与状态映射

### 默认管理根

- `~/CursorProjects`
- `~/DeepseekProjects`

可在配置中增删。

### 每根下的生命周期容器

```text
<root>/
  _pm/
    incubator/   # idea
    archive/     # archived
    trash/       # trash（不真删）
  <project>/     # active | paused | kept
```

根目录只暴露一个 `_pm`，避免 `_incubator` / `_archive` / `_trash` 并列增生。

### 与旧目录的兼容

若仍存在根级 `_incubator` / `_archive`：

- `scan` **识别为等价位置**，按路径启发状态
- 新的 `set` 搬动一律写入 `_pm/...`
- 一次性迁入 `_pm` 的迁移命令不在 MVP 必做范围

### 状态模型

**主状态（生命周期，六选一）**

| 状态 | 物理位置 | 含义 |
|------|----------|------|
| `idea` | `<root>/_pm/incubator/<name>` | 想法 / 孵化中 |
| `active` | `<root>/<name>` | 在研或在用 |
| `paused` | `<root>/<name>` | 故意搁置，以后可能捡起 |
| `kept` | `<root>/<name>` | 不开发但有参考/复用价值 |
| `archived` | `<root>/_pm/archive/<name>` | 退出日常视野，留底 |
| `trash` | `<root>/_pm/trash/<name>` | 待删隔离，不自动真删 |

**辅标记**

- `purpose: known | unknown`
- `unknown` 不是最终态：补上用途后应落到六个主状态之一
- 写入非空 `summary` 时自动将 `purpose` 设为 `known`

### 新发现时的默认状态

| 路径启发 | 默认 status | 默认 purpose |
|----------|-------------|--------------|
| 在 `_pm/incubator` 或旧 `_incubator` | `idea` | `unknown` |
| 在 `_pm/archive` 或旧 `_archive` | `archived` | `unknown` |
| 在 `_pm/trash` | `trash` | `unknown` |
| 日常区（根下一层） | `paused` | `unknown` |

默认不用 `active`，避免把沉睡项目标成在研。

### `set` 搬目录规则

- 先打印计划（from → to）
- 默认需确认；`--yes` 跳过确认；`--dry-run` 只打印不执行
- 目标路径已存在则中止并报错
- 只搬已登记的项目根，不碰 `_pm` 容器本身
- MVP **不支持跨 root 搬迁**

### 什么算「一个项目」

- 仅扫描配置 root 下的**一层**子目录
- 跳过：`_pm`、以 `.` 开头的目录、配置 `ignore` 列表
- 不把 monorepo 子包拆成多条登记（整仓一条）

## 3. 数据模型与命令面

### 配置

路径：`~/.config/bay/config.yaml`

建议字段：

- `roots: []string` — 绝对路径
- `ignore: []string` — 目录名忽略
- `stale_after_days: int` — 默认如 `90`，供 `doctor` 使用
- `pm_dir: string` — 默认 `_pm`

### 登记库

路径：`~/.local/share/bay/ledger.db`

表 `projects`（字段可微调，语义固定）：

| 字段 | 说明 |
|------|------|
| `id` | 主键 |
| `name` | 目录名 |
| `root` | 所属配置根（绝对路径） |
| `path` | 项目当前绝对路径 |
| `status` | 六个主状态之一 |
| `purpose_known` | bool |
| `summary` | 一句话用途，可空 |
| `signals` | JSON 文本：扫描信号 |
| `user_set_status` | bool；曾被 `set --status` 则锁定，scan 不以路径启发覆盖 |
| `missing` | bool；磁盘上找不到 |
| `last_seen_at` | 最近一次 scan 见到 |
| `last_moved_at` | 最近一次 bay 搬目录 |
| `created_at` / `updated_at` | 元时间 |

约束：

- `(root, name)` 唯一
- `path` 唯一

### MVP 命令

| 命令 | 作用 |
|------|------|
| `bay scan` | 扫 roots：新增登记 / 更新 path 与 signals / 标记 missing |
| `bay ls` | 列表；过滤 `--status`、`--unknown`、`--root`、关键词 |
| `bay show <name>` | 单条详情 |
| `bay set <name> --status …` / `--summary "…"` | 改状态或用途；状态变化触发位置搬迁（见上） |
| `bay doctor` | 巡检报告；`--md <file>` 可输出 Markdown |

### 写库策略（`scan`）

- **新目录**：插入；`summary` 空；`purpose=unknown`；status 按路径启发；`user_set_status=false`
- **已登记**：更新 `path`、`signals`、`last_seen_at`；清除或设置 `missing`
- **不覆盖**：已有 `summary`；若 `user_set_status=true` 则不覆盖 `status`
- **库有盘无**：标记 `missing=true`，不删除登记行

### 消歧与退出码

- 跨 root 同名：`show` / `set` 要求 `--root` 或等价消歧
- 退出码：成功 `0`；用户取消确认、无匹配等非 `0`，便于脚本

### 明确留到以后

- `export`（YAML/Markdown 导出）
- 真删命令
- `bay pin`（项目内摘要文件）
- 跨根搬迁、自动改名、TUI
- 旧 `_incubator` / `_archive` 一键迁入 `_pm`

## 4. 扫描信号、`doctor` 规则与风险

### 扫描信号（只存库，不自动当作 summary）

- 目录名
- `README.md` / `README`：首个非空标题或前约 500 字；空则记 `readme: empty`
- 若有 `.git`：最后提交时间、最近一条 commit subject
- 目录 mtime（无 git 时的活跃度兜底）
- 浅特征标签：是否存在 `package.json` / `go.mod` / `pyproject.toml` / `Cargo.toml` 等（不解析依赖树）

### `doctor` 规则（只建议，不自动 set / 不自动搬）

1. `purpose=unknown` → 优先补用途
2. `status=paused` 且「最后活跃时间」早于 `stale_after_days` → 建议 `archived` 或 `trash`（最后活跃时间 = signals 中 git 最后提交时间与目录 mtime 的较新者；皆无则用 `last_seen_at`）
3. `status=active` 同样久未动 → 提示是否改为 `paused`
4. `status=trash` → 列出待手动清理
5. `missing=true` → 列出失踪项
6. 报告默认打印终端；`--md` 可落盘

### 风险与护栏

| 风险 | 护栏 |
|------|------|
| 误搬目录 | 默认确认 + `--dry-run`；目标存在则中止 |
| 跨根同名 | 强制消歧 |
| 误收非项目目录 | 只扫一层 + ignore；误收可 `set … trash` |
| 库/配置丢失 | 单文件库，用户自行备份；MVP 无云同步 |
| 手挪目录与库不一致 | 下次 `scan` 校正 `path`；仅当 `user_set_status=false` 时可用路径启发刷新 status |

### 验收标准（第一版成功）

对两个默认 root 跑通：

1. `bay scan` 入库  
2. `bay ls` / `bay show` 可查  
3. `bay set` 改状态并（在确认后）搬入 `_pm/...`  
4. `bay doctor` 产出：未知用途、陈旧候选、trash 清单、失踪项  

## 5. 实现路径选择（已定）

采用「精简港湾」：

- Go 单二进制
- SQLite 为项目登记唯一真相
- YAML 仅作配置（非项目主数据双写）
- 不做智能用途推断优先；扫描信号供人参考

配置与数据默认路径：

- 配置：`~/.config/bay/config.yaml`
- 库：`~/.local/share/bay/ledger.db`

## 6. 开放问题（实现计划阶段再钉）

- Go module 路径 / 仓库远程是否在立项时创建
- `stale_after_days` 默认值是否 90
- `bay ls` 默认排序（名称 vs 最近活跃）
- 旧 `_incubator`（本设计文档所在仓）与未来 `_pm/incubator` 的迁移时机
