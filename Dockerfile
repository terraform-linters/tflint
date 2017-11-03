FROM alpine:3.5

MAINTAINER wata727

COPY dist/linux_amd64/tflint /usr/local/bin

ENTRYPOINT ["tflint"]

WORKDIR /data
