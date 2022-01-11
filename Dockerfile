FROM golang:1.16-alpine3.15 AS builder

ARG ALPINE_PKG_BASE="make git"
ARG ALPINE_PKG_EXTRA=""

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache ${ALPINE_PKG_BASE} ${ALPINE_PKG_EXTRA}

WORKDIR /edge

COPY . .

RUN make update || echo "skipping"
RUN make build

FROM alpine:3.15

WORKDIR /edge

COPY --from=builder /edge/edge-device-manager ./edge-device-manager
COPY --from=builder /edge/bin ./bin
COPY --from=builder /edge/etc ./etc
COPY --from=builder /edge/public ./public

RUN ./bin/boot.sh

EXPOSE 10996

CMD ["./edge-device-manager", "-cp", "etc", "-cn", "config"]

