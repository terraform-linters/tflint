FROM golang:1.18.3-alpine3.15 as builder

RUN apk add --no-cache make

WORKDIR /tflint
COPY . /tflint
RUN make build

FROM alpine:3.16.0 as prod

LABEL maintainer=terraform-linters

RUN apk add --no-cache ca-certificates

COPY --from=builder /tflint/dist/tflint /usr/local/bin

ENTRYPOINT ["tflint"]
WORKDIR /data
