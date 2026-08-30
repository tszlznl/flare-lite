FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
# 模板与样式通过 go:embed 编进二进制，产物不再依赖源码目录
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags "-s -w" -o /out/nav .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /data
COPY --from=build /out/nav /usr/local/bin/nav

EXPOSE 5120
# 数据文件写在 WORKDIR 下，挂 volume 即可持久化
VOLUME ["/data"]
ENTRYPOINT ["/usr/local/bin/nav"]
