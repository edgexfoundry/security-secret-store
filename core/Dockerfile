FROM alpine:3.7

RUN mkdir -p /edgex/res

WORKDIR /edgex

COPY res/configuration-docker.toml res/configuration.toml

ADD edgex-vault-worker .

ENTRYPOINT ["./edgex-vault-worker"]

CMD  ["--init=true"]