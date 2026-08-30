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

WORKDIR /src

# 依赖层单独缓存，改源码不会重新下载模块
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags "-s -w" -o /out/nav .

# ---------- 运行阶段 ----------
# nonroot 变体以 uid 65534 运行；如需以 root 跑（省掉数据目录权限问题），
# 构建时传 --build-arg DISTROLESS_TAG=latest。
FROM ${DISTROLESS_REPO}:${DISTROLESS_TAG}

COPY --from=build --chown=65534:65534 /out/nav /nav

# 时区数据由 main.go 里的 `_ "time/tzdata"` 内嵌进二进制，
# 所以这里只需给个默认值，运行期 -e TZ=... 即可覆盖，不必重新构建。
ENV TZ=Asia/Shanghai
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
