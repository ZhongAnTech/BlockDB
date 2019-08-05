# Build OG from alpine based golang environment
FROM golang:1.12-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ENV GOPROXY https://goproxy.io
ENV GO111MODULE on

WORKDIR /go/src/github.com/annchain/BlockDB
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make blockdb

# Copy OG into basic alpine image
FROM alpine:latest

RUN apk add --no-cache curl iotop busybox-extras tzdata

COPY --from=builder /go/src/github.com/annchain/BlockDB/deployment/config.toml /opt/config.toml
COPY --from=builder /go/src/github.com/annchain/BlockDB/build/blockdb /opt/

# for a temp running folder. This should be mounted from the outside
RUN mkdir /rw

EXPOSE 28017 28018 28019

WORKDIR /opt

CMD ["./blockdb", "--config", "/opt/config.toml", "--multifile_by_level", "--log_line_number", "--log_dir", "/rw/log/", "--datadir", "/rw/datadir", "run"]



