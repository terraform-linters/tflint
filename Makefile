default: build

prepare:
	cd tools; go run ./plugin-stub-gen; cd ../

test: prepare
	go test -timeout 5m $$(go list ./... | grep -v test-fixtures | grep -v vendor | grep -v aws-sdk-go)

build: test
	mkdir -p dist
	go build -v -o dist/tflint

install: test
	go install

lint:
	golint --set_exit_status $$(go list ./...)
	go vet ./...

clean:
	rm -rf dist/

code: prepare
	go generate ./...

tools:
	go install github.com/golang/mock/mockgen
	go install golang.org/x/lint/golint

.PHONY: prepare test build install lint clean code tools
