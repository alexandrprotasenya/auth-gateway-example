FROM golang:1.17.3-alpine
WORKDIR /build
ADD go.mod .
ADD go.sum .
RUN go mod download

ADD main.go .
RUN go build -o app main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /build/app ./
CMD ["./app"]
