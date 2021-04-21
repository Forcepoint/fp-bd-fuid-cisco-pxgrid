FROM golang:alpine as builder
RUN mkdir /build
ADD . /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -a -installsuffix cgo -ldflags '-extldflags "-static"' -o fuid-ise .
FROM scratch
FROM alpine:3.12
# install openssl
RUN apk add --update openssl && \
    apk add --no-cache bash && \
    rm -rf /var/cache/apk/*
COPY --from=builder /build/fuid-ise $GOPATH/bin
RUN export GODEBUG=x509ignoreCN=0
WORKDIR /app
CMD ["bash"]
