BINARY = usg
VERSION = 0.1.0
LDFLAGS = -s -w

.PHONY: build clean all

build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) .

all: build-darwin-arm64 build-linux-arm64 build-linux-amd64

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 .

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-arm64 .

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 .

clean:
	rm -f $(BINARY)
	rm -rf dist/
