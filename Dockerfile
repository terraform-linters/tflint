FROM golang:1.15-alpine3.12 as builder

RUN apk add --no-cache make

WORKDIR /tflint
COPY . /tflint
RUN make build

FROM alpine:3.12 as prod

LABEL maintainer=terraform-linters

RUN apk add --no-cache ca-certificates

COPY --from=builder /tflint/dist/tflint /usr/local/bin

ENTRYPOINT ["tflint"]
WORKDIR /data
