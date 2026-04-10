# isetup

[English](README.md)

跨平台 CLI 工具，自动检测操作系统、硬件架构，然后自适应调用正确的安装命令，一键部署你的开发环境。

**仅限命令行工具。** isetup 专为命令行工程师设计，默认模板只包含终端工具，不涉及任何 GUI 应用。

新机器？`isetup install`，搞定。

## 特性

- **自动检测** — 操作系统、CPU 架构、发行版、GPU、可用包管理器、Shell
- **全平台支持** — macOS、Linux（Ubuntu/Fedora/Arch）、Windows、WSL
- **Profile 分组** — 按用途组织工具（`lang-runtimes`、`base`、`git-tools`、`python-dev`、`ai-tools`、`gpu`、`shell-enhancements`、`system-tools`）
- **自适应安装** — 根据系统自动选择 `brew`、`apt`、`choco`、`winget`、`dnf`、`pacman` 或自定义脚本
- **Root / Docker 感知** — 自动检测 UID 0 并省略 `sudo`，在没有安装 `sudo` 的容器内也能正常工作
- **模板变量** — Shell 命令中支持 `{{.Arch}}`、`{{.OS}}`、`{{.Home}}`，实现架构感知下载
- **依赖排序** — `depends_on` 确保工具按正确顺序安装
- **条件安装** — `when: has_gpu` 在没有 GPU 的机器上自动跳过
- **跳过已安装** — 自动检测 PATH 中已有的工具并跳过，`-f` 强制重装
- **丰富诊断** — 完整的命令输出、环境快照、耗时记录，存储在 `~/.isetup/logs/`
- **实时进度** — `[N/Total]` 计数器 + 系统信息头，不会静默等待
- **试运行模式** — 预览所有命令，不实际执行

## 安装

**Linux / macOS（推荐）：**

```bash
curl -fsSL https://raw.githubusercontent.com/host452b/isetup/main/install.sh | bash
```

自定义安装目录：

```bash
INSTALL_DIR=~/.local/bin curl -fsSL https://raw.githubusercontent.com/host452b/isetup/main/install.sh | bash
```

**Windows (PowerShell)：**

```powershell
$version = (Invoke-RestMethod "https://api.github.com/repos/host452b/isetup/releases/latest").tag_name.TrimStart('v')
$url = "https://github.com/host452b/isetup/releases/download/v$version/isetup_${version}_windows_amd64.zip"
Invoke-WebRequest $url -OutFile isetup.zip
Expand-Archive isetup.zip -DestinationPath .
Move-Item -Force isetup.exe "$env:USERPROFILE\AppData\Local\Microsoft\WindowsApps\"
Remove-Item isetup.zip
```

**Go install：**

```bash
go install github.com/host452b/isetup@latest
```

**从源码构建：**

```bash
git clone https://github.com/host452b/isetup.git && cd isetup && go build -o isetup .
```

**验证：**

```bash
isetup version
```

## 快速开始

```bash
# 生成默认配置（可选 — 没有配置文件也能用内置默认模板）
isetup init

# 编辑你的配置
vim ~/.isetup.yaml

# 预览将要安装的内容
isetup install --dry-run

# 安装所有工具
isetup install

# 仅安装指定 profile
isetup install -p base,ai-tools
```

## 默认工具

内置模板安装 **61 个工具**，分布在 8 个 profile 中：

### lang-runtimes — 语言运行时与版本管理

| 工具 | 说明 |
|------|------|
| nvm | Node.js 版本管理器 — 按项目切换 Node 版本 |
| node-lts | Node.js LTS 版本，通过 nvm 安装 |
| typescript | TypeScript 编译器（tsc） |
| golang | Go 编程语言 |
| rust | Rust 工具链 via rustup（rustc、cargo、rustfmt） |
| miniconda | Conda 包/环境管理器 — 默认不激活 base，需手动 `conda activate` |
| mise | 多语言运行时管理器（asdf 替代品，管理 Node/Python/Go 等版本） |

### base — 核心命令行工具

**编辑器 & 终端**

| 工具 | 说明 |
|------|------|
| git | 分布式版本控制 |
| neovim | 现代终端编辑器（Vim 分支） |
| tmux | 终端复用器 — 分屏、会话保持 |
| tmux-ide | 脚本化 tmux 会话布局（npm） |

**搜索 & 导航**

| 工具 | 说明 |
|------|------|
| fzf | 模糊搜索 — 交互式过滤文件、历史命令、分支 |
| ripgrep | 递归搜索替代品 (rg)，极速 |
| fd | 更简单快速的 `find` 替代，默认设置更合理 |
| tree | 树形可视化目录结构 |

**现代 CLI 替代品（Rust 驱动）**

| 工具 | 说明 |
|------|------|
| bat | `cat` 替代 — 语法高亮 + Git 集成 |
| eza | `ls` 替代 — 图标、颜色、Git 状态 |

**数据处理**

| 工具 | 说明 |
|------|------|
| jq | 命令行 JSON 处理器 — 解析 API 响应、转换配置 |
| yq | YAML/TOML/XML 处理器 — 编辑 CI 配置、K8s 清单 |

**系统工具**

| 工具 | 说明 |
|------|------|
| htop | 交互式进程查看器 |
| btop | 资源监控 TUI（CPU、内存、磁盘、网络） |
| make | 构建自动化 — 很多仓库默认就有 Makefile |
| curl | URL 数据传输（HTTP 客户端、API 测试） |
| wget | 文件下载器，支持断点续传 |
| zip | 压缩工具 |
| unzip | 解压工具 |
| fonts-firacode | Fira Code 编程字体，带连字 |

### git-tools — Git & CI/CD

| 工具 | 说明 |
|------|------|
| gh | GitHub CLI — 终端管理 PR、Issue、Actions |
| glab | GitLab CLI — 终端管理 MR、Pipeline、Issue |
| lazygit | Git 终端 UI — 可视化暂存、分支、变基 |
| delta | Git diff 分页器 — 语法高亮 + 并排对比 |
| gitlab-runner | GitLab CI 运行器 — 本地运行 CI 任务 |

### python-dev — Python 生态

| 工具 | 说明 |
|------|------|
| uv | 超快 Python 包安装器（pip 替代） |
| pip-tools | httpie（HTTP 客户端）、black（格式化）、ruff（lint） |
| pip-build-tools | build、twine、hatchling — Python 包发布工具 |
| huggingface-hub | Hugging Face CLI — 下载/上传模型和数据集 |
| pr-analyzers | gitlab-pr-analyzer、github-pr-analyzer、jira-lens |
| playwright | 浏览器自动化测试（Chromium/Firefox/WebKit） |
| pgcli | PostgreSQL CLI — 自动补全 + 语法高亮 |
| ai-ml-libs | chromadb（向量数据库）、pgvector、langsmith、langfuse（LLM 可观测性） |

### ai-tools — AI & LLM

| 工具 | 说明 |
|------|------|
| claude-code | Anthropic Claude Code — 终端 AI 编程助手 |
| codex-cli | OpenAI Codex CLI — AI 代码生成 |
| cursor | Cursor AI 编辑器（CLI 安装器） |
| yoyo | AI 代理自动审批 PTY 代理 |
| arxs | 多源学术论文搜索 CLI |
| ollama | 本地运行 LLM（Llama、Mistral 等） |

### gpu — NVIDIA GPU（条件: `when: has_gpu`）

| 工具 | 说明 |
|------|------|
| cuda-toolkit | NVIDIA CUDA 编译器和库 |
| nvidia-driver | NVIDIA 驱动 v550 |

### shell-enhancements — Shell 效率增强

| 工具 | 说明 |
|------|------|
| zoxide | 智能 `cd` — 学习你最常用的目录 |
| starship | 跨 Shell 提示符 — 显示 Git 状态、语言版本，配置极简 |
| direnv | 按目录自动加载 `.envrc` — 每项目管理环境变量 |

### system-tools — 调试 & 网络

| 工具 | 说明 |
|------|------|
| lsof | 列出打开的文件和端口 — 查找占用 8080 端口的进程 |
| netcat | TCP/UDP 瑞士军刀 — 测试连接、端口扫描 |
| tcpdump | 网络抓包分析 |
| dnsutils | DNS 查询工具：`dig`、`nslookup` |
| strace | 系统调用追踪 — 调试进程行为（仅 Linux） |
| sqlite3 | SQLite 数据库 CLI — 轻量数据库查询调试 |

## 命令一览

```
isetup init                      生成默认 ~/.isetup.yaml
isetup init --force              覆盖已有配置
isetup detect                    输出系统检测信息（JSON）
isetup install                   安装所有 profile
isetup install -p base,ai-tools  安装指定 profile
isetup install -f                强制重装已安装的工具
isetup install --dry-run         仅预览命令，不执行
isetup install --timeout 5m     设置单工具超时（默认 10 分钟）
isetup list                      列出所有 profile 和工具
isetup version                   打印版本号
```

## 配置指南

配置文件位于 `~/.isetup.yaml`（可用 `--config` 指定其他路径）。

运行 `isetup init` 生成默认模板后，按照你的需求编辑即可。

### 基本结构

```yaml
version: 1                    # 配置格式版本，固定为 1

settings:
  log_level: info              # 日志级别：debug / info / warn / error
  dry_run: false               # 设为 true 则仅打印命令不执行

profiles:
  profile-name:                # 自定义 profile 名称
    when: has_gpu              # 可选：条件，不满足则跳过整组
    tools:
      - name: tool-name        # 工具名称（必填，全局唯一）
        depends_on: other-tool # 可选：依赖另一个工具，确保先安装
        apt: package-name      # 可选：apt 包名
        brew: package-name     # 可选：brew 包名
        choco: package-name    # 可选：choco 包名
        # ... 更多安装方式见下方
```

### 如何定义一个工具

每个工具可以声明多种安装方式，isetup 运行时根据当前系统自动选择：

```yaml
- name: git                    # 必填：工具名
  apt: git                     # Linux (Debian/Ubuntu)
  dnf: git                     # Linux (Fedora/RHEL)
  pacman: git                  # Linux (Arch)
  brew: git                    # macOS
  choco: git                   # Windows（优先）
  winget: Git.Git              # Windows（备选）
```

### 使用 shell 自定义安装脚本

对于没有包管理器支持的工具，使用 `shell` 字段：

```yaml
- name: nvm
  shell:
    # 精确匹配 OS（优先级最高）
    linux: "curl -o- https://example.com/install.sh | bash"
    darwin: "brew install nvm"
    windows: "irm https://example.com/install.ps1 | iex"

    # unix 是 linux + darwin 的 fallback
    unix: "curl -o- https://example.com/install.sh | bash"

# 也支持字符串简写形式（仅 unix）
- name: tmux
  shell: "curl -fsSL https://tmux.example.com/install.sh | sh"
```

**Shell 匹配优先级：**
1. `shell.linux` / `shell.darwin` / `shell.windows` — 精确 OS 匹配
2. `shell.unix` — linux + darwin 通用 fallback
3. `shell: "字符串"` — unix-only 简写

### 使用 pip 和 npm

```yaml
# pip 安装（优先使用 conda pip，fallback 到 pip3/pip）
- name: python-linters
  depends_on: miniconda       # 建议先安装 miniconda
  pip:
    - ruff
    - black
    - httpie

# npm 全局安装（需要 Node.js 已安装）
- name: codex-cli
  npm: "@openai/codex"
```

### 模板变量

Shell 命令支持 Go 模板语法，运行时自动替换为当前系统值：

```yaml
- name: miniconda
  shell:
    linux: |
      curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-{{.Arch}}.sh -o /tmp/miniconda.sh
      bash /tmp/miniconda.sh -b -p $HOME/miniconda3
```

| 变量 | 说明 | 示例值 |
|------|------|--------|
| `{{.Arch}}` | 平台原生架构标签 | `x86_64`、`aarch64`（Linux arm64）、`arm64`（macOS） |
| `{{.OS}}` | 操作系统 | `linux`、`darwin`、`windows` |
| `{{.Distro}}` | 发行版 | `Ubuntu 22.04.3 LTS`、`macOS 15.3` |
| `{{.Home}}` | 用户主目录 | `/home/user`、`/Users/user` |

### 依赖管理

使用 `depends_on` 确保安装顺序：

```yaml
- name: nvm
  shell:
    unix: "curl -o- https://nvm.sh/install.sh | bash"

- name: node-lts
  depends_on: nvm             # nvm 安装完成后才会安装 node-lts
  shell:
    unix: "source ~/.nvm/nvm.sh && nvm install --lts"
```

规则：
- 依赖是**全局跨 profile** 的 — 工具 A 可以依赖另一个 profile 中的工具 B
- 如果依赖的工具未被选中安装（如 `-p` 过滤掉了），则跳过并警告
- 如果依赖的工具安装失败，下游工具也会跳过

### 条件安装

目前支持的条件：

```yaml
profiles:
  gpu:
    when: has_gpu              # 仅在检测到 GPU 时安装
    tools:
      - name: cuda-toolkit
        apt: nvidia-cuda-toolkit
```

| 条件 | 含义 |
|------|------|
| `has_gpu` | 检测到 NVIDIA GPU |

### 配置校验规则

isetup 在安装前自动校验配置：

| 规则 | 级别 |
|------|------|
| 缺少 `version` 字段或版本不支持 | 错误（终止） |
| 工具名重复（跨 profile） | 错误（终止） |
| 工具缺少 `name` 字段 | 错误（终止） |
| 循环依赖 | 错误（终止） |
| 未知的 `when` 条件 | 错误（终止） |
| `depends_on` 引用不存在的工具 | 警告（运行时跳过） |
| Profile 的 `tools` 列表为空 | 警告 |

### 包管理器命令展开表

| 配置键 | 展开为 | 提权 |
|--------|--------|------|
| `apt: X` | `sudo apt-get install -y X` | sudo（root 时省略） |
| `dnf: X` | `sudo dnf install -y X` | sudo（root 时省略） |
| `pacman: X` | `sudo pacman -S --noconfirm X` | sudo（root 时省略） |
| `brew: X` | `brew install X` | 无 |
| `choco: X` | `choco install X -y` | 需提权 shell |
| `winget: X` | `winget install --id X -e --accept-source-agreements` | 无 |
| `pip: [X, Y]` | `pip3 install X Y`（conda pip 优先） | 无 |
| `npm: X` | `npm install -g X` | 无 |

### 完整配置示例

```yaml
version: 1

settings:
  log_level: info
  dry_run: false

profiles:
  base:
    tools:
      - name: git
        apt: git
        dnf: git
        pacman: git
        brew: git
        choco: git

      - name: neovim
        apt: neovim
        brew: neovim
        choco: neovim

      - name: tmux
        apt: tmux
        brew: tmux

      - name: fzf
        apt: fzf
        brew: fzf
        choco: fzf

      - name: ripgrep
        apt: ripgrep
        brew: ripgrep
        choco: ripgrep

  lang-runtimes:
    tools:
      - name: nvm
        shell:
          unix: |
            curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash

      - name: node-lts
        depends_on: nvm
        shell:
          unix: "source ~/.nvm/nvm.sh && nvm install --lts"

      - name: golang
        brew: go
        shell:
          linux: |
            GO_VERSION=$(curl -fsSL "https://go.dev/VERSION?m=text" | head -1)
            curl -fsSL "https://go.dev/dl/${GO_VERSION}.linux-{{.Arch}}.tar.gz" -o /tmp/go.tar.gz
            sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tar.gz

      - name: rust
        shell:
          unix: "curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y"

  python-dev:
    tools:
      - name: miniconda
        shell:
          linux: |
            curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-{{.Arch}}.sh -o /tmp/miniconda.sh
            bash /tmp/miniconda.sh -b -p $HOME/miniconda3
            $HOME/miniconda3/bin/conda config --set auto_activate_base false
            $HOME/miniconda3/bin/conda init
          darwin: |
            curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-MacOSX-{{.Arch}}.sh -o /tmp/miniconda.sh
            bash /tmp/miniconda.sh -b -p $HOME/miniconda3
            $HOME/miniconda3/bin/conda config --set auto_activate_base false
            $HOME/miniconda3/bin/conda init
          windows: |
            irm https://repo.anaconda.com/miniconda/Miniconda3-latest-Windows-x86_64.exe -OutFile miniconda.exe
            Start-Process .\miniconda.exe -ArgumentList '/S /D=%USERPROFILE%\miniconda3' -Wait

      - name: pip-tools
        depends_on: miniconda
        pip:
          - httpie
          - black
          - ruff

  ai-tools:
    tools:
      - name: claude-code
        shell:
          unix: "curl -fsSL https://claude.ai/install.sh | bash"
          windows: "irm https://claude.ai/install.ps1 | iex"

      - name: codex-cli
        npm: "@openai/codex"

      - name: cursor
        shell:
          unix: "curl -fsS https://cursor.com/install | bash"

      - name: yoyo
        shell:
          unix: "curl -fsSL https://github.com/host452b/yoyo/releases/latest/download/install.sh | sh"

      - name: arxs
        shell:
          unix: "curl -fsSL https://raw.githubusercontent.com/host452b/arxs/main/install.sh | sh"
          windows: "irm https://raw.githubusercontent.com/host452b/arxs/main/install.ps1 | iex"

  gpu:
    when: has_gpu
    tools:
      - name: cuda-toolkit
        apt: nvidia-cuda-toolkit

      - name: nvidia-driver
        apt: nvidia-driver-550
```

## 日志与诊断

日志目录：`~/.isetup/logs/`（可用 `--log-dir` 覆盖）

每次运行产生两个文件：

- `isetup-<时间戳>.env.json` — 完整环境快照（OS、架构、GPU、PATH 等）
- `isetup-<时间戳>.log` — 逐工具安装记录（命令、stdout、stderr、退出码、耗时）

终端输出示例：

```
[✓] git              (brew  ) 0.8s
[✓] neovim           (brew  ) 3.2s
[✓] nvm              (shell ) 2.1s
[✗] cuda-toolkit     (apt   ) 1.1s  → see log for details
─────────────────────────────
Installed: 3 | Failed: 1 | Skipped: 0
Log: ~/.isetup/logs/isetup-2026-03-24T20-57-30.log
```

安装失败时，日志包含完整的 stdout/stderr 输出，方便 AI 或人工回溯定位问题。

## 系统检测

`isetup detect` 以 JSON 格式输出完整系统信息：

```json
{
  "os": "darwin",
  "arch": "arm64",
  "arch_label": "arm64",
  "distro": "macOS 15.3.2",
  "kernel": "24.3.0",
  "wsl": false,
  "is_root": false,
  "shell": "/bin/zsh",
  "gpu": {
    "detected": true,
    "model": "Apple M3 Pro"
  },
  "pkg_managers": ["brew", "pip3", "npm"]
}
```

## 项目结构

```
isetup/
├── main.go                  # 入口
├── embed.go                 # 嵌入默认模板
├── cmd/                     # CLI 命令（cobra）
├── internal/
│   ├── config/              # YAML 解析 + 校验
│   ├── detector/            # OS/GPU/Shell/包管理器 检测
│   ├── executor/            # 安装引擎（解析器、执行器、拓扑排序）
│   └── logger/              # 结构化日志
└── template/
    └── default.yaml         # 默认配置模板
```

## 故障排查

**工具安装卡住：**
使用 `--timeout 2m` 设置更短的单工具超时。检查 `~/.isetup/logs/` 获取完整命令输出。

**权限被拒绝：**
isetup 自动检测 root 并省略 `sudo`。如果你不是 root 且 `sudo` 失败，请确保你的用户有 sudo 权限。

**找不到 Profile：**
使用 `isetup list` 查看可用的 profile。Profile 名称区分大小写。

**已安装的工具：**
isetup 会跳过 PATH 中已有的工具。使用 `-f` 强制重装。

**自定义配置路径：**
使用 `--config /path/to/config.yaml` 指定非默认配置文件。

## 退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 所有工具安装或跳过成功 |
| 1 | 一个或多个工具安装失败 |
| 2 | 配置错误（无效 YAML、校验失败） |

## 构建

```bash
go build -o isetup .
```

需要 Go 1.22+。

## PowerShell 兼容性

Windows 上同时兼容 PowerShell 5.1 和 PowerShell 7+：
- 优先使用 `pwsh`（7+）
- Fallback 到 `powershell.exe`（5.1）
- `irm | iex` 在两个版本上都可用

## License

MIT
