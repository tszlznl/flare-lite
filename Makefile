.PHONY: run debug build test vet fmt docker docker-multiarch clean

APP     := flare-lite
PORT    ?= 25000
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
	docker build -t $(APP) .

docker-multiarch:
	docker buildx build --platform linux/amd64,linux/arm64 -t $(APP) . --push

clean:
	rm -rf bin
