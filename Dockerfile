FROM golang:1.12-alpine as builder

ENV GOPROXY https://goproxy.cn
ENV GO111MODULE on

ADD . /BlockDB
WORKDIR /BlockDB
RUN make blockdb


FROM base-registry.zhonganinfo.com/env/traefik-alpline:3.9.4
WORKDIR /
COPY --from=builder BlockDB/config.toml .
COPY --from=builder BlockDB/build/blockdb .
EXPOSE 28017 28018 28019 8080
CMD ["./blockdb", "--config", "config.toml", "-n", "run"]