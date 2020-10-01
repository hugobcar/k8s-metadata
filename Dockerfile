FROM golang:1.14 as builder
WORKDIR /go/src/github.com/hugobcar/k8s-metadata
ADD . /go/src/github.com/hugobcar/k8s-metadata
# RUN GO111MODULE=on go mod vendor
RUN CGO_ENABLED=0 go build -o k8s-metadata

FROM alpine:3.12.0
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY --from=builder /go/src/github.com/hugobcar/k8s-metadata/k8s-metadata .
CMD ["./k8s-metadata"]