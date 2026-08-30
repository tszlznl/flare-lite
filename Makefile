.PHONY: run debug build test vet fmt docker docker-debug docker-multiarch clean

APP     := flare-lite
PORT    ?= 5120
TZ      ?= Asia/Shanghai
LDFLAGS := -s -w

run:
	go run . -port $(PORT)

debug:
	go run . -port $(PORT) -debug

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(APP) .

test:
	go test -timeout 120s ./...

vet:
	go vet ./...
	gofmt -l .

fmt:
	gofmt -w .

docker:
	docker build --build-arg TZ=$(TZ) -t $(APP) .

# 带 busybox shell 的 distroless 变体，用于进容器排查
docker-debug:
	docker build --build-arg DISTROLESS_TAG=debug-nonroot -t $(APP):debug .

docker-multiarch:
	docker buildx build --platform linux/amd64,linux/arm64 \
		--build-arg TZ=$(TZ) -t $(APP) . --push

clean:
	rm -rf bin
