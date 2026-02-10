.PHONY: build test install clean

build:
	cd go && go build -o bin/rig ./cmd/rig

test:
	cd go && go test ./...

install:
	cd go && go install ./cmd/rig

clean:
	rm -rf go/bin

help:
	@echo "Available targets:"
	@echo "  build    - Build the rig binary to go/bin/rig"
	@echo "  test     - Run all tests"
	@echo "  install  - Install rig to \$$GOPATH/bin"
	@echo "  clean    - Remove build artifacts"
