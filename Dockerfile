# syntax=docker/dockerfile:1

# 容器打包：以 distroless 静态镜像为运行时基础。
# 本应用 CGO_ENABLED=0 全静态编译、且运行期不发起任何外部请求，
# 因此用最小的 static 变体即可（约 2 MiB，比 alpine 还小一半）。

ARG GO_VERSION=1.26
ARG DISTROLESS_REPO=gcr.io/distroless/static-debian13
ARG DISTROLESS_TAG=nonroot

# ---------- 构建阶段 ----------
# --platform=$BUILDPLATFORM 让构建始终跑在本机架构上交叉编译，
# 交叉编译出多架构产物，比在 QEMU 下跑 go build 快得多。
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build

ARG TARGETOS
ARG TARGETARCH
ARG TZ=Asia/Shanghai

WORKDIR /src

# tzdata 只为把时区文件取出来放进最终镜像，见下方说明
RUN apk add --no-cache tzdata

# 依赖层单独缓存，改源码不会重新下载模块
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags "-s -w" -o /out/nav .

# distroless 镜像里没有 /usr/share/zoneinfo，而本应用用 time.Now() 渲染
# {date} 这类占位符并打印日志时间戳。缺时区数据时 Go 会静默退回 UTC，
# {date} 就会在早上 8 点（东八区）而不是午夜翻篇。
# 这里只取需要的那一个时区文件，避免把整个 zoneinfo（约 10 MB）塞进 2 MB 的镜像。
RUN mkdir -p "/z/$(dirname "${TZ}")" && cp "/usr/share/zoneinfo/${TZ}" "/z/${TZ}"

# ---------- 运行阶段 ----------
# nonroot 变体以 uid 65534 运行；如需以 root 跑（省掉数据目录权限问题），
# 构建时传 --build-arg DISTROLESS_TAG=latest。
FROM ${DISTROLESS_REPO}:${DISTROLESS_TAG}

ARG TZ=Asia/Shanghai

COPY --from=build --chown=65534:65534 /out/nav /nav
COPY --from=build /z/ /usr/share/zoneinfo/

ENV TZ=${TZ}
# 数据文件写在 WORKDIR 下，挂 volume 即可持久化。
# 注意：用 bind mount 挂空目录时，该目录需对 uid 65534 可写，
# 否则首次运行生成 sites.yml 会失败（命名卷会自动继承镜像内属主，无此问题）。
WORKDIR /data

EXPOSE 5120
VOLUME ["/data"]

# distroless 没有 shell，ENTRYPOINT/CMD 必须用向量形式，
# 否则运行时会试图拼接一个不存在的 shell 而启动失败。
ENTRYPOINT ["/nav"]
CMD ["-port", "5120"]
