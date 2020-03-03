FROM golang:1.14.0-alpine3.11 as builder

RUN apk --no-cache add git make gcc musl-dev zip

WORKDIR /tflint
ADD . /tflint
RUN make build

FROM alpine:3.10 as prod

LABEL maintainer=terraform-linters

RUN apk add --no-cache ca-certificates

COPY --from=builder /tflint/dist/tflint /usr/local/bin

ENTRYPOINT ["tflint"]
WORKDIR /data
