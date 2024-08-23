default: build

prepare:
	git submodule init
	git submodule update
	go run ./plugin/stub-generator

test: prepare
	go test -timeout 5m $$(go list ./... | grep -v test-fixtures | grep -v stub-generator | grep -v integrationtest)

build:
	mkdir -p dist
	go build -v -o dist/tflint

wasi:
	GOOS=wasip1 GOARCH=wasm go build -mod=vendor -o main.wasm main.go

debug-wasi:
	make wasi && wasmtime --env TFLINT_LOG=debug --dir=../vanguard-demo/wordpress/infra/aws::/demo main.wasm --chdir /demo --recursive

install:
	go install

e2e: prepare install
	go test -timeout 5m $$(go list ./integrationtest/... | grep -v race)

e2e-race: prepare
	go test --race --timeout 5m ./integrationtest/race

lint:
	golangci-lint run ./...
	cd terraform/ && golangci-lint run ./...

clean:
	rm -rf dist/

generate:
	go generate ./...

release:
	go run ./tools/release/main.go

.PHONY: prepare test build install e2e lint clean generate
