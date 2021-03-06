FROM golang:1.13 as build
WORKDIR /go/src/github.com/jgadling/pennant/
ADD . /go/src/github.com/jgadling/pennant/
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=build /go/src/github.com/jgadling/pennant/pennant /bin/pennant
ADD configs/pennant.json /etc/
CMD ["/bin/pennant", "server", "-c", "/etc/pennant.json"]
