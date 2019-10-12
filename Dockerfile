FROM golang:1.13 AS builder

ENV GOARCH 386
ENV GOOS linux

WORKDIR /work

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .
RUN go build -o unifi-exporter

FROM alpine:3.7

COPY --from=builder /work/unifi-exporter .

ENTRYPOINT ["./unifi-exporter"]
