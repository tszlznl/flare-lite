# flare-lite

> 极简高效的个人导航站 —— 基于 Go + Echo v5，无数据库、零 JavaScript、单 YAML 驱动、资源内嵌、单文件轻量部署。

[![build](https://github.com/tszlznl/flare-lite/actions/workflows/build.yml/badge.svg)](https://github.com/tszlznl/flare-lite/actions/workflows/build.yml)
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL%20v3-blue.svg)](./LICENSE)
[![Docker Image](https://img.shields.io/badge/Docker-GHCR-blue?logo=docker)](https://github.com/tszlznl/flare-lite/pkgs/container/flare-lite)

---

## ✨ 核心特性

- ⚡ **极致轻量**：纯 Go 全静态编译，单二进制独立运行；内存占用低（通常 < 15MB）。
- 📝 **单文件数据源**：所有书签与站点配置均保存在单个 `sites.yml` 中。修改即刻生效，无需重启服务。
- 🔒 **零 JavaScript & 强安全**：原生服务端 HTML 渲染，自带 `Content-Security-Policy: script-src 'none'`，完全杜绝前端脚本注入风险。
- 🐳 **Alpine 运行时**：Docker 镜像基于 Alpine Linux 构建，内置标准 Shell 及基础工具，方便排查与日常维护。
- 🌐 **国际化域名 (IDN)**：支持中文及非 ASCII 域名（如 `琳.tw`），自动且标准地转换为 Punycode（`xn--jgy.tw`）。
- 📅 **动态日期占位符**：URL 支持 `{year}`, `{month}`, `{day}`, `{date}`, `{compact-date}`，方便收藏每日一变的历史归档或日报地址。
- 🎨 **资源内嵌**：HTML 模板与 CSS 样式通过 `go:embed` 打包在二进制中，开箱即用。

---

## 🚀 快速开始

### 方式一：Docker Compose（推荐）

1. 创建 `docker-compose.yml`（仓库根目录下已内置）：

```yaml
services:
  flare-lite:
    image: ghcr.io/tszlznl/flare-lite:latest
    container_name: flare-lite
    restart: unless-stopped
    ports:
      - "25000:25000"
    environment:
      - TZ=Asia/Shanghai
      - NAV_PORT=25000
    volumes:
      - ./data:/data
```

2. 启动服务：

```bash
docker compose up -d
```

3. 访问 **http://localhost:25000** 即可。首次启动会自动在挂载的 `./data` 目录下生成默认的 `sites.yml` 示例。

---

### 方式二：Docker 单命令运行

```bash
docker run -d \
  --name flare-lite \
  -p 25000:25000 \
  -e TZ=Asia/Shanghai \
  -v flare-lite-data:/data \
  --restart unless-stopped \
  ghcr.io/tszlznl/flare-lite:latest
```

> **提示**：如果使用宿主机目录挂载（如 `-v ./data:/data`），因为镜像以非 root 用户（uid:gid `65534:65534`）运行，初次挂载空目录时请确保该目录对 uid 65534 有写权限（`sudo chown -R 65534:65534 ./data`）。使用 Docker 命名卷时会自动处理属主，无此要求。

---

### 方式三：源码 / 本地二进制运行

确保已安装 Go 1.22+ 环境：

```bash
# 运行（首次启动在当前目录自动创建 sites.yml）
go run .

# 换端口
go run . -port 8080

# 调试模式（修改 embed/templates 或 embed/assets 下文件后刷新即可见，免去重新编译）
go run . -debug
```

编译独立二进制可执行文件：

```bash
# Linux / macOS
go build -trimpath -ldflags "-s -w" -o bin/nav .

# Windows (PowerShell)
go build -trimpath -ldflags "-s -w" -o bin/nav.exe .
```

产物不依赖任何外部动态链接库与模板文件，将编译生成的 `nav` 与 `sites.yml` 放置在同一目录即可运行。

---

## ⚙️ 命令行参数与环境变量

可以通过命令行参数或对应的环境变量控制服务行为。**命令行参数优先级高于环境变量**：

| 命令行参数 | 对应环境变量 | 默认值 | 说明 |
| :--- | :--- | :--- | :--- |
| `-port` | `NAV_PORT` | `25000` | 监听端口，可填写纯端口号（`25000`）或完整地址（`0.0.0.0:25000` / `127.0.0.1:25000`） |
| `-config` | `NAV_CONFIG` | `sites.yml` | 书签数据文件路径（YAML 格式） |
| `-debug` | `NAV_DEBUG` | `false` | 调试模式：设为 `1` 或 `true` 时，模板与静态样式将直接读取本地磁盘文件 |

---

## 📄 数据配置指南 (`sites.yml`)

完整示例可参考仓库中的 [sites.example.yml](./sites.example.yml)。

```yaml
site:
  title: 連結收藏               # 站点标题
  footer: ""                   # 底部文本；留空则自动显示「共 N 條連結」或搜索结果统计
  open_in_new_tab: true        # 链接是否在新标签页中打开
  show_search: true            # 是否展示顶部搜索框

# 可选：声明分组。配置后 links 会按 group 分割为多张表格；不声明则统一平铺在单张表展示
groups:
  - id: linux
    title: Linux 與自架
  - id: life
    title: 生活與閱讀

links:
  - name: 烤雞堡的筆記
    link: https://wei.dev
    desc: 討論 self-hosting 技術與雲端管理。
    group: linux               # 对应 groups 中的 id（可选）

  - name: 琳的備忘手札
    link: https://琳.tw        # 支持中文域名，自动处理为标准 punycode
    desc: Linux 企業運維知識。
    group: linux

  - name: 今日日報
    link: https://example.com/report/{compact-date} # 20260831
    desc: 支持日期占位符。
    group: life
```

### 动态占位符

在 `link` 字段中支持使用以下占位符，访问时会根据当前系统/容器时区自动替换：

| 占位符 | 替换示例 | 说明 |
| :--- | :--- | :--- |
| `{year}` | `2026` | 4 位年份 |
| `{month}` | `08` | 2 位月份（补零） |
| `{day}` | `31` | 2 位日期（补零） |
| `{date}` | `2026-08-31` | 标准日期格式（ISO 8601） |
| `{compact-date}` | `20260831` | 紧凑无分隔符日期 |

---

## 📂 项目结构

```text
├── cmd/                     # 命令行参数与环境变量解析
├── config/
│   ├── data/                # YAML 读写与基于 mtime 的轻量级内存缓存
│   ├── define/              # 全局配置定义与路径规范化
│   └── model/               # 数据结构定义
├── embed/                   # 编译期内嵌资源
│   ├── assets/css/          # 样式文件 (style.css)
│   └── templates/           # 页面模板 (index.html)
├── internal/
│   ├── fn/                  # 辅助函数：URL 补全、动态日期占位符、RFC 3492 Punycode
│   ├── pages/home/          # 首页渲染：关键词过滤、分组逻辑与表格生成
│   ├── resources/           # 模板引擎与静态资源路由处理
│   └── server/              # Echo v5 服务路由组装与生命周期管理
├── Dockerfile               # 多阶段交叉编译与 Alpine 运行时镜像构建
├── docker-compose.yml       # Docker Compose 编排文件
├── sites.example.yml        # 示例配置文件
└── main.go                  # 程序入口
```

---

## 🛠️ Makefile 快捷指令

```bash
make run              # 本地启动（监听 25000 端口）
make debug            # 启动调试模式
make build            # 静态编译二进制产物至 bin/flare-lite
make test             # 执行单元测试
make vet              # 代码静态检查与格式验证
make fmt              # 自动格式化 Go 代码
make docker           # 本地构建 Docker 镜像
make clean            # 清理编译生成文件
```

---

## 📜 许可证与鸣谢

- 本项目采用 **[AGPL-3.0](./LICENSE)** 许可证。
- 整体架构与设计思路参考 [soulteary/flare](https://github.com/soulteary/flare)（AGPL-3.0），表格 UI 启发自社区精美博客书签分享。
