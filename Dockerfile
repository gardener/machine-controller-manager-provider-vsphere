#############      builder                                  #############
FROM golang:1.19.2 AS builder

WORKDIR /go/src/github.com/gardener/machine-controller-manager-provider-vsphere
COPY . .

RUN .ci/build

#############      base                                     #############
FROM gcr.io/distroless/static-debian11:nonroot AS base


#############      machine-controller               #############
FROM base AS machine-controller
WORKDIR /

COPY --from=builder /go/src/github.com/gardener/machine-controller-manager-provider-vsphere/bin/rel/machine-controller /machine-controller
ENTRYPOINT ["/machine-controller"]
