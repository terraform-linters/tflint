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

image:
	docker build -t wata727/tflint:${VERSION} .
	docker tag wata727/tflint:${VERSION} wata727/tflint:latest
	docker push wata727/tflint:${VERSION}
	docker push wata727/tflint:latest

rule:
	go run tools/rule_generator.go

model_rules:
	go run github.com/wata727/tflint/tools/model-rule-gen

.PHONY: default prepare test build install release clean code image rule
