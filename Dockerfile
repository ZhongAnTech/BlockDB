FROM golang:1.13-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers git

ENV GOPROXY https://goproxy.cn
ENV GO111MODULE on

#ADD . /BlockDB
#WORKDIR /BlockDB
#RUN make blockdb
#
#
#FROM alpine:latest
#WORKDIR /
#COPY --from=builder BlockDB/config.toml .
#COPY --from=builder BlockDB/build/blockdb .

WORKDIR /go/src/github.com/annchain/BlockDB
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN make blockdb

# Copy OG into basic alpine image
FROM alpine:latest

RUN apk add --no-cache curl iotop busybox-extras tzdata

WORKDIR /
COPY --from=builder /go/src/github.com/annchain/BlockDB/deployment/config.toml .
COPY --from=builder /go/src/github.com/annchain/BlockDB/blockdb .

# for a temp running folder. This should be mounted from the outside
RUN mkdir /rw

EXPOSE 28017 28018 28019 8080
CMD ["./blockdb", "--config", "config.toml", "-n", "run"]