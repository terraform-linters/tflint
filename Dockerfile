FROM golang:1.13.1-alpine3.10 as builder

RUN apk --no-cache add git make gcc musl-dev zip

ENV GO111MODULE=on

WORKDIR /tflint
ADD . /tflint
RUN make build

FROM alpine:3.10 as prod

LABEL maintainer=terraform-linters

RUN apk add --no-cache ca-certificates

COPY --from=builder /tflint/dist/tflint /usr/local/bin

ENTRYPOINT ["tflint"]
WORKDIR /data
