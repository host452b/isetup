# isetup

[English](README.md)

跨平台 CLI 工具，自动检测操作系统、硬件架构，然后自适应调用正确的安装命令，一键部署你的开发环境。

**仅限命令行工具。** isetup 专为命令行工程师设计，默认模板只包含终端工具，不涉及任何 GUI 应用。

新机器？`isetup install`，搞定。

## 特性

- **自动检测** — 操作系统、CPU 架构、发行版、GPU、可用包管理器、Shell
- **全平台支持** — macOS、Linux（Ubuntu/Fedora/Arch）、Windows、WSL
- **Profile 分组** — 按用途组织工具（`base`、`node-dev`、`python-dev`、`ai-tools`、`gpu`）
- **自适应安装** — 根据系统自动选择 `brew`、`apt`、`choco`、`winget`、`dnf`、`pacman` 或自定义脚本
- **模板变量** — Shell 命令中支持 `{{.Arch}}`、`{{.OS}}`、`{{.Home}}`，实现架构感知下载
- **依赖排序** — `depends_on` 确保工具按正确顺序安装
- **条件安装** — `when: has_gpu` 在没有 GPU 的机器上自动跳过
- **丰富诊断** — 完整的命令输出、环境快照、耗时记录，存储在 `~/.isetup/logs/`
- **试运行模式** — 预览所有命令，不实际执行

## 安装

**一行命令安装（Linux / macOS）：**

```bash
curl -fsSL https://github.com/host452b/isetup/releases/latest/download/isetup_$(uname -s | tr '[:upper:]' '[:lower:]')_$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/').tar.gz | tar xz -C /usr/local/bin isetup
```

**Windows (PowerShell)：**

```powershell
irm "https://github.com/host452b/isetup/releases/latest/download/isetup_0.1.0_windows_amd64.zip" -OutFile isetup.zip; Expand-Archive isetup.zip -DestinationPath .; Move-Item isetup.exe $env:USERPROFILE\AppData\Local\Microsoft\WindowsApps\; Remove-Item isetup.zip
```

**Go install：**

```bash
go install github.com/isetup-dev/isetup@latest
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

## 命令一览

```
isetup init                      生成默认 ~/.isetup.yaml
isetup init --force              覆盖已有配置
isetup detect                    输出系统检测信息（JSON）
isetup install                   安装所有 profile
isetup install -p base,node-dev  安装指定 profile
isetup install --dry-run         仅预览命令，不执行
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
| `apt: X` | `sudo apt-get install -y X` | sudo |
| `dnf: X` | `sudo dnf install -y X` | sudo |
| `pacman: X` | `sudo pacman -S --noconfirm X` | sudo |
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

  node-dev:
    tools:
      - name: nvm
        shell:
          unix: |
            curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.7/install.sh | bash
          windows: |
            irm https://github.com/coreybutler/nvm-windows/releases/download/1.1.12/nvm-setup.exe -OutFile nvm-setup.exe

      - name: node-lts
        depends_on: nvm
        shell:
          unix: "source ~/.nvm/nvm.sh && nvm install --lts"
          windows: "nvm install lts && nvm use lts"

  python-dev:
    tools:
      - name: miniconda
        shell:
          linux: |
            curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-{{.Arch}}.sh -o /tmp/miniconda.sh
            bash /tmp/miniconda.sh -b -p $HOME/miniconda3
            $HOME/miniconda3/bin/conda init
          darwin: |
            curl -fsSL https://repo.anaconda.com/miniconda/Miniconda3-latest-MacOSX-{{.Arch}}.sh -o /tmp/miniconda.sh
            bash /tmp/miniconda.sh -b -p $HOME/miniconda3
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
