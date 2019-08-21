default: build

prepare:
	go mod vendor

test: prepare
	go test $$(go list ./... | grep -v vendor | grep -v aws-sdk-go)

build: test
	mkdir -p dist
	go build -v -o dist/tflint

install: test
	go install

release: test
	goreleaser --rm-dist

clean:
	rm -rf dist/

code: prepare
	go generate ./...

.PHONY: default prepare test build install release clean code
