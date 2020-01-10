FROM alpine:3.10

COPY bin/rel/cmi-plugin /cmi-plugin
WORKDIR /
ENTRYPOINT ["/cmi-plugin"]
