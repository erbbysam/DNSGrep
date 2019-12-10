FROM golang:1.8 as builder

WORKDIR /go/src/dnsgrep
COPY . .

RUN go get "github.com/jessevdk/go-flags"
RUN go get "github.com/gorilla/mux"
RUN go get "github.com/golang/example/stringutil"
RUN go build -v -o dnsgrep .

FROM alpine
RUN adduser -S -D -H -h /app user
USER user
COPY --from=builder /go/src/dnsgrep /dnsgrep
WORKDIR /dnsgrep
ENTRYPOINT ["/dnsgrep/dnsgrep"]
