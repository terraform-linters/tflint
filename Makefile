default: build

prepare:
	go get -u github.com/golang/dep/cmd/dep
	dep ensure -v

test: prepare
	go test $$(go list ./... | grep -v vendor | grep -v mock)

build: test
	mkdir -p dist
	go build -v -o dist/tflint

install: test
	go install

release: test
	go get github.com/mitchellh/gox
	gox --output 'dist/{{.OS}}_{{.Arch}}/{{.Dir}}'
	mkdir -p dist/releases
	zip -j dist/releases/tflint_darwin_386.zip    dist/darwin_386/tflint
	zip -j dist/releases/tflint_darwin_amd64.zip  dist/darwin_amd64/tflint
	zip -j dist/releases/tflint_freebsd_386.zip   dist/freebsd_386/tflint
	zip -j dist/releases/tflint_freebsd_amd64.zip dist/freebsd_amd64/tflint
	zip -j dist/releases/tflint_freebsd_arm.zip   dist/freebsd_arm/tflint
	zip -j dist/releases/tflint_linux_386.zip     dist/linux_386/tflint
	zip -j dist/releases/tflint_linux_amd64.zip   dist/linux_amd64/tflint
	zip -j dist/releases/tflint_linux_arm.zip     dist/linux_arm/tflint
	zip -j dist/releases/tflint_netbsd_386.zip    dist/netbsd_386/tflint
	zip -j dist/releases/tflint_netbsd_amd64.zip  dist/netbsd_amd64/tflint
	zip -j dist/releases/tflint_netbsd_arm.zip    dist/netbsd_arm/tflint
	zip -j dist/releases/tflint_openbsd_386.zip   dist/openbsd_386/tflint
	zip -j dist/releases/tflint_openbsd_amd64.zip dist/openbsd_amd64/tflint
	zip -j dist/releases/tflint_windows_386.zip   dist/windows_386/tflint.exe
	zip -j dist/releases/tflint_windows_amd64.zip dist/windows_amd64/tflint.exe

clean:
	rm -rf dist/

mock: prepare
	go generate ./...
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/ec2/ec2iface/interface.go -destination mock/ec2mock.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/elasticache/elasticacheiface/interface.go --destination mock/elasticachemock.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/elb/elbiface/interface.go -destination mock/elbmock.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/elbv2/elbv2iface/interface.go -destination mock/elbv2mock.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/iam/iamiface/interface.go -destination mock/iammock.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/rds/rdsiface/interface.go -destination mock/rdsmock.go -package mock
	mockgen -source vendor/github.com/aws/aws-sdk-go/service/ecs/ecsiface/interface.go  -destination mock/ecsmock.go -package mock

image:
	docker build -t wata727/tflint:${VERSION} .
	docker tag wata727/tflint:${VERSION} wata727/tflint:latest
	docker push wata727/tflint:${VERSION}
	docker push wata727/tflint:latest

rule:
	go run tools/rule_generator.go

.PHONY: default prepare test build install release clean mock image rule
