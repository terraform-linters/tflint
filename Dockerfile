FROM golang:1.11.5-alpine3.9 as builder

MAINTAINER wata727

WORKDIR /root

ENV GOPATH /root/.go
ENV PATH $GOPATH/bin:$PATH

RUN apk add --no-cache --update git ca-certificates make gcc g++
RUN go get -d github.com/wata727/tflint

WORKDIR /root/.go/src/github.com/wata727/tflint

RUN make build

FROM alpine:3.9

MAINTAINER wata727

COPY --from=builder /root/.go/src/github.com/wata727/tflint/tflint /usr/local/bin/tflint

CMD ["ash"]
