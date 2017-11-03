FROM alpine:3.5

MAINTAINER wata727

RUN apk add --no-cache ca-certificates

COPY dist/linux_amd64/tflint /usr/local/bin

ENTRYPOINT ["tflint"]
