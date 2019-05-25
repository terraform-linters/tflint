FROM golang:1.12-alpine as builder

RUN apk --no-cache add git make gcc musl-dev zip

WORKDIR /go/src/github.com/wata727/tflint/

ADD . /go/src/github.com/wata727/tflint

RUN make build

FROM alpine:3.9 as prod

LABEL maintainer=wata727

RUN apk add --no-cache ca-certificates

COPY --from=builder /go/src/github.com/wata727/tflint/dist/tflint /usr/local/bin

ENTRYPOINT ["tflint"]

WORKDIR /data
