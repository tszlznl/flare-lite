# syntax=docker/dockerfile:1

# 容器打包：以 Alpine Linux 镜像为运行时基础。
# 本应用 CGO_ENABLED=0 全静态编译，Alpine 体积小巧且自带常用工具与 Shell。

ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.21

# ---------- 构建阶段 ----------
# --platform=$BUILDPLATFORM 让构建始终跑在本机架构上交叉编译，
# 交叉编译出多架构产物，比在 QEMU 下跑 go build 快得多。
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS build

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src

# 依赖层单独缓存，改源码不会重新下载模块
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags "-s -w" -o /out/nav .

# ---------- 运行阶段 ----------
FROM alpine:${ALPINE_VERSION}

# 安装基础 CA 证书与时区包
RUN apk --no-cache add ca-certificates tzdata

COPY --from=build /out/nav /nav

# 时区默认东八区，运行期可通过 -e TZ=... 覆盖
ENV TZ=Asia/Shanghai

# 数据文件写在 WORKDIR 下，挂 volume 即可持久化
WORKDIR /data

EXPOSE 25000
VOLUME ["/data"]

ENTRYPOINT ["/nav"]
CMD ["-port", "25000"]
