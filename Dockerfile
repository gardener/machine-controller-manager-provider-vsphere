#############      builder                                  #############
FROM eu.gcr.io/gardener-project/3rd/golang:1.16.11 AS builder

WORKDIR /go/src/github.com/gardener/machine-controller-manager-provider-vsphere
COPY . .

RUN .ci/build

#############      base                                     #############
FROM eu.gcr.io/gardener-project/3rd/alpine:3.13.5 AS base

RUN apk add --update bash curl tzdata
WORKDIR /

#############      machine-controller               #############
FROM base AS machine-controller

COPY --from=builder /go/src/github.com/gardener/machine-controller-manager-provider-vsphere/bin/rel/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
