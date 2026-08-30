.PHONY: run debug build test vet fmt docker clean

APP     := nav
PORT    ?= 5120
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

clean:
	rm -rf bin
