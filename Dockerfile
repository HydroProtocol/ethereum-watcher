FROM golang:1.13

WORKDIR /go/src
COPY . /go/src 

RUN go build -o bin/ethereum-watcher cli/main.go

FROM alpine
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN apk --no-cache add ca-certificates

COPY --from=0 /go/src/bin/* /bin/

CMD ["/bin/ethereum-watcher"]
