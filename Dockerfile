#############      builder                                  #############
FROM eu.gcr.io/gardener-project/3rd/golang:1.17.5 AS builder

WORKDIR /go/src/github.com/gardener/machine-controller-manager-provider-vsphere
COPY . .

RUN .ci/build

#############      base                                     #############
FROM eu.gcr.io/gardener-project/3rd/alpine:3.15.0 AS base

RUN apk add --update bash curl tzdata
WORKDIR /

#############      machine-controller               #############
FROM base AS machine-controller

COPY --from=builder /go/src/github.com/gardener/machine-controller-manager-provider-vsphere/bin/rel/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
