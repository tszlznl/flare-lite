# 轻量导航站 —— 技术栈对齐 soulteary/flare：Go + Echo v5，无数据库，资源内嵌，单文件部署。

`sites.yml` 就是全部数据。改完不用重启，下一次请求即生效。

## 快速开始

```bash
go run .                # 首次运行自动生成示例 sites.yml，监听 5120
go run . -port 8080     # 换端口
go run . -debug         # 调试模式：模板/样式改完刷新即生效，无需重编译
```

浏览器打开 http://localhost:5120

编译成单个可执行文件：

```bash
go build -trimpath -ldflags "-s -w" -o bin/nav.exe .
```

产物不依赖任何外部文件，把 `nav` 和 `sites.yml` 放一起就能跑。

## 命令行参数

| 参数       | 环境变量       | 默认值       | 说明                             |
| ---------- | -------------- | ------------ | -------------------------------- |
| `-port`    | `NAV_PORT`     | `5120`       | 端口，或写完整的 `host:port`     |
| `-config`  | `NAV_CONFIG`   | `sites.yml`  | 数据文件路径                     |
| `-debug`   | `NAV_DEBUG=1`  | `false`      | 模板与样式改为读磁盘             |

## 数据格式

```yaml
site:
  title: 連結收藏
  footer: ""                # 留空则自动显示「共 N 條連結」/ 搜索结果统计
  open_in_new_tab: true
  show_search: true

# 可选：声明后 links 会按 group 拆成多张表，不声明就是一张平铺长表
groups:
  - id: linux
    title: Linux 與自架

links:
  - name: 烤雞堡的筆記       # 名稱
    link: https://wei.dev    # URL
    desc: 討論 self-hosting 技術與雲端管理。   # 第一印象
    group: linux             # 可选
```

`link` 可以省掉协议（`www.zaqizaba.xyz`），渲染时自动补 `https://`，
但 URL 列仍按原样显示。也支持日期占位符，适合收藏每日一变的地址：

```yaml
  - name: 今日日報
    link: https://example.com/report/{compact-date}   # 20260830
```

可用占位符：`{year}` `{month}` `{day}` `{date}` `{compact-date}`。

## 项目结构

```
├── cmd/                     命令行与环境变量解析
├── config/
│   ├── data/                YAML 读写 + 基于 mtime 的进程内缓存
│   ├── define/              全局启动参数与路径解析
│   └── model/               数据结构，与 YAML 一一对应
├── embed/                   编译期内嵌的模板与样式
│   ├── templates/index.html
│   └── assets/css/style.css
├── internal/
│   ├── fn/                  URL 规范化、动态占位符、punycode
│   ├── pages/home/          首页渲染：过滤、分组、组装表格
│   ├── resources/           模板渲染器 + 静态资源路由
│   └── server/              Echo 组装与启动
└── main.go
```

## 设计取舍

- **零 JavaScript**：响应头带 `Content-Security-Policy: script-src 'none'`，
  搜索走 `<form method="get">`，全部逻辑在服务端。
- **无数据库**：一个 YAML 文件即数据源，靠 mtime + size 判断是否失效，
  不做 watch、不做常驻缓存刷新任务。
- **模板自动转义**：数据里的任意内容都经 `html/template` 转义，
  不用手工拼接 HTML 字符串。
- **国际化域名**：`琳.tw` 这类域名在写进 `href` 前会转成 punycode
  （`xn--jgy.tw`）。`html/template` 默认会对非 ASCII host 做百分号编码，
  那不是合法的 host 写法。punycode 编码按 RFC 3492 自行实现，
  期望值以 .NET `IdnMapping` 为基准做了对拍验证。

## 待打磨

当前是可直接跑通的 demo，以下是明确还没做、等你测完再定的部分：

- 图标：参考图没有图标列，暂未加。可接 favicon 自动抓取或本地图标集。
- 编辑界面：现在只能改 YAML 文件，flare 有个内置 CSV 编辑器，可参考。
- 名称列可点击性：参考图里「名稱」是纯文本、只有 URL 是链接，
  当前严格照图实现。若要整行可点需要调整表格语义。
- 空简介目前显示 `—`，参考图是留白，一行 CSS 即可改回。
- 深色模式、访问计数、链接存活检测。

## Docker

运行时基础镜像用 [distroless](https://github.com/GoogleContainerTools/distroless) 的
`static-debian13`：本应用 `CGO_ENABLED=0` 全静态编译、运行期不发起任何外部请求，
所以不需要 glibc、也不需要 CA 证书。

```bash
docker build -t flare-lite .
docker run -d -p 5120:5120 --name flare-lite flare-lite
```

多架构（NAS / 树莓派常见）：

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t flare-lite . --push
```

### 时区

distroless 镜像内没有 `/usr/share/zoneinfo`，而本应用用 `time.Now()` 渲染
`{date}` 占位符并打印日志时间戳。缺时区数据时 Go 会**静默退回 UTC**，
`{date}` 就会在早上 8 点（东八区）而不是午夜翻篇。

Dockerfile 默认把 `Asia/Shanghai` 的单个时区文件拷进镜像，可换：

```bash
docker build --build-arg TZ=Asia/Taipei -t flare-lite .
```

### 数据持久化

镜像以 uid `65534`（nonroot）运行，`/data` 是工作目录，`sites.yml` 生成在这里。

- 命名卷会自动继承镜像内目录属主，直接挂即可：
  `docker run -d -p 5120:5120 -v flare-lite-data:/data flare-lite`
- 用 bind mount 挂**空宿主目录**时，该目录需要对 uid 65534 可写，
  否则首次生成 `sites.yml` 会失败：`sudo chown 65534 ./data`
- 嫌麻烦可退回 root 运行：`docker build --build-arg DISTROLESS_TAG=latest -t flare-lite .`

### 调试

distroless 没有 shell，`docker exec -it <c> sh` 是用不了的。官方提供带 busybox 的
`:debug` 变体，构建时换标签即可：

```bash
docker build --build-arg DISTROLESS_TAG=debug-nonroot -t flare-lite:debug .
docker run --entrypoint=sh -ti flare-lite:debug
```

## 许可证

AGPL-3.0，见 [LICENSE](./LICENSE)。

注意 AGPL 第 13 条：如果你修改后通过网络提供服务（导航站正是这种形态），
必须向用户提供修改后的完整源码。个人自托管一般不受影响，但公开部署前值得了解。

## 来源

架构与技术栈参考 [soulteary/flare](https://github.com/soulteary/flare)（AGPL-3.0）：
Go + Echo v5、YAML 文件存储、`go:embed` 内嵌资源、零 JS 服务端渲染。

本仓库为独立实现，未复制 flare 的源码，但目录分层与设计思路受其影响。
表格 UI 取自一份博客收藏列表的截图。

容器打包思路来自[《使用以语言为中心的容器基础镜像 distroless》](https://soulteary.com/2021-10-14/use-language-centric-container-base-image-distroless.html)（2021）。
该文的部分写法按现行 distroless 已经需要调整，本仓库的 Dockerfile 按官方现状实现，
差异见 [Docker](#docker) 一节。
