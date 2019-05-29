FROM golang:1.12-alpine as builder

RUN apk --no-cache add git make gcc musl-dev zip

ENV GO111MODULE=on

WORKDIR /tflint
ADD . /tflint
RUN make build

FROM alpine:3.9 as prod

LABEL maintainer=wata727

RUN apk add --no-cache ca-certificates

COPY --from=builder /tflint/dist/tflint /usr/local/bin

ENTRYPOINT ["tflint"]
WORKDIR /data
