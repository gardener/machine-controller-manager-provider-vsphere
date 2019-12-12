# machine-controller-manager-provider-vsphere
Out of tree (gRPC based) implementation for `vSphere` as a new provider.

## About
This repository implements a CMI plugin for managing machines (VMs) on vSphere. It implements the services defined at [machine-spec](https://github.com/gardener/machine-spec).

Details about the ProviderSpec and the Secrets can be found in [provider_spec.go](./pkg/vsphere/apis/provider_spec.go)

## Prerequisites

- vSphere installation, user with permissions to create/delete VMs in the specified data center
- A vSphere network with an DHCP server. For Gardener, the network is created by the vsphere infrastructure
  controller, which needs VMware NSX-T to setup the software-defined network, SNAT and DHCP.
- Suitable VM templates must already be deployed on vSphere. The provider uses the `guestId` to identify the
  correct way to initiate a cloud-init. 
  Supported OS are
  - CoreOS images using igniton for cloud-init (see https://stable.release.core-os.net/amd64-usr for images)
    In this case make sure, that the `guestId` is overwritten with `coreos64Guest` in the ProviderSpec.    
  - Other Linux cloud images with a cloud-init VApp (e.g. Ubuntu at https://cloud-images.ubuntu.com/releases)
    can be used if they meet the requirements like Docker, SystemD, ... (see 
    [Gardener contract for OperationSystemConfig](https://github.com/gardener/gardener/blob/master/docs/extensions/operatingsystemconfig.md) 
    for more details)

## Supports vSphere versions

Currently this provider has only be tested with vSphere 6.7 (Update 3)

