FROM --platform=$BUILDPLATFORM golang:1.19-alpine3.16 as builder

ARG TARGETOS TARGETARCH

RUN apk add --no-cache make

WORKDIR /tflint
COPY . /tflint
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH make build

FROM alpine:3.17.1

LABEL maintainer=terraform-linters

RUN apk add --no-cache ca-certificates

COPY --from=builder /tflint/dist/tflint /usr/local/bin

ENTRYPOINT ["tflint"]
WORKDIR /data
