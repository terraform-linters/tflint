FROM golang:1.15-alpine3.12 as builder

RUN apk --no-cache add git make gcc musl-dev zip

WORKDIR /tfenv
ARG TFENV_VERSION=v2.0.0
RUN git clone https://github.com/tfutils/tfenv.git . && git checkout --quiet ${TFENV_VERSION}

WORKDIR /tflint
ADD . /tflint
RUN make build

FROM alpine:3.12 as prod

LABEL maintainer=terraform-linters

RUN apk add --no-cache ca-certificates curl bash

COPY --from=builder /tflint/dist/tflint /usr/local/bin

COPY --from=builder /tfenv /tfenv
RUN ln -s /tfenv/bin/* /usr/local/bin

ENTRYPOINT ["tflint"]
WORKDIR /data
