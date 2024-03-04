VERSION=1.4.0
OUTPUT=_local/pkg
.PHONY: build_linux build_macos pkg_linux pkg_macos all default clean setup test

default: test

clean:
	@rm -rf $(OUTPUT)

test:
	@go test ./...

setup:
	@mkdir -p $(OUTPUT)

build_linux: setup
	@env GOOS=linux GOARCH=amd64 go build -o $(OUTPUT) -ldflags "-X main.version=$(VERSION)" .

build_macos: setup
	@env GOOS=darwin GOARCH=arm64 go build -o $(OUTPUT) -ldflags "-X main.version=$(VERSION)" .

pkg_linux: build_linux
	@tar -C $(OUTPUT) -czf $(OUTPUT)/enpeeem$(VERSION).linux-amd64.tar.gz enpeeem

pkg_macos: build_macos
	@tar -C $(OUTPUT) -czf $(OUTPUT)/enpeeem$(VERSION).macos-arm64.tar.gz enpeeem

all: clean test pkg_linux pkg_macos
