/*
Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package api

const (
	// TagClusterPrefix is the old tag prefix for tagging the cluster name
	TagClusterPrefix = "kubernetes.io/cluster/"
	// TagNodeRolePrefix is the old tag prefix for tagging the node role
	TagNodeRolePrefix = "kubernetes.io/role/"
	// TagMCMClusterName is the tag key for tagging a VM with the cluster name
	TagMCMClusterName = "mcm.gardener.cloud/cluster"
	// TagMCMRole is the tag key for tagging a VM with its role (e.g 'node')
	TagMCMRole = "mcm.gardener.cloud/role"

	// LabelMCMVSphere is the tag key for labeling a VM as a managed machine
	LabelMCMVSphere = "mcm.gardener.cloud/machine"

	// VSphereKubeconfig is the key for the kubeconfig in the credentials secret
	VSphereKubeconfig = "vsphereKubeconfig"
)

// VsphereProviderSpec is an interface to hide the concrete spec version
type VsphereProviderSpec struct {
	V1 *VsphereProviderSpec1 `json:"v1,omitempty"`
	V2 *VsphereProviderSpec2 `json:"v2,omitempty"`

	// SSHKeys is an optional array of ssh public keys to deploy to VM (may already be included in UserData)
	// +optional
	SSHKeys []string `json:"sshKeys,omitempty"`
	// Tags to be placed on the VM
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// VsphereProviderSpec1 contains the fields of
// provider spec that the plugin expects
type VsphereProviderSpec1 struct {
	// Region is the vSphere region
	Region string `json:"region"`

	// Datacenter is the vSphere data center
	Datacenter string `json:"datacenter"`

	// DatastoreCluster is the data store cluster to use (either DatastoreCluster or Datastore must be specified)
	// +optional
	DatastoreCluster string `json:"datastoreCluster,omitempty"`
	// Datastore is the data store to use (either DatastoreCluster or Datastore must be specified)
	// +optional
	Datastore string `json:"datastore,omitempty"`

	// ComputeCluster is the compute cluster to use for placement (either ComputeCluster, ResourcePool, or HostSystem must be specified)
	// +optional
	ComputeCluster string `json:"computeCluster,omitempty"`
	// ResourcePool is the resource pool to use  for placement (either ComputeCluster, ResourcePool, or HostSystem must be specified)
	// +optional
	ResourcePool string `json:"resourcePool,omitempty"`
	// HostSystem is the host system to use for placement (either ComputeCluster, ResourcePool, or HostSystem must be specified)
	// +optional
	HostSystem string `json:"hostSystem,omitempty"`

	// Folder is the folder to place VMs into
	// +optional
	Folder string `json:"folder,omitempty"`
	// NumCpus is the number of virtual CPUs of the VM
	NumCpus int `json:"numCpus"`
	// Memory is VM memory size in MB
	Memory int `json:"memory"`
	// MemoryReservationLockedToMax is flag to reserve all guest OS memory (no swapping in ESXi host)
	// +optional
	MemoryReservationLockedToMax *bool `json:"memoryReservationLockedToMax,omitempty"`
	// SystemDisk specifies the system disk
	// +optional
	SystemDisk *VSphereSystemDisk `json:"systemDisk,omitempty"`
	// ExtraConfig allows to specify additional VM options.
	// e.g. sched.swap.vmxSwapEnabled=false to disable the VMX process swap file
	// +optional
	ExtraConfig map[string]string `json:"extraConfig,omitempty"`
	// Network is the vSphere network to use
	Network string `json:"network"`
	// SwitchUUID is VDS UUID (only needed if there are multiple virtual distributed switches the network is assigned to)
	// +optional
	SwitchUUID string `json:"switchUuid"`
	// TemplateVM is the VM template to clone
	TemplateVM string `json:"templateVM"`
	// GuestID is an optional value to overwrite the VM guest id of the templae
	// +optional
	GuestID string `json:"guestId,omitempty"`
	// VApp contains the Properties of the VApp to start on booting
	VApp *VApp `json:"vapp,omitempty"`
	// Force is an experimental flag to overwrite an existing VM with the same name
	// +optional
	Force bool `json:"force,omitempty"`
	// WaitForIP is an experimental flag if controller should wait until VM has IP assigned
	// +optional
	WaitForIP bool `json:"waitForIP,omitempty"`
	// Customization is an experimental option to add a CustomizationSpec
	// +optional
	Customization string `json:"customization,omitempty"`
}

// SpecVersion returns spec version
func (s *VsphereProviderSpec1) SpecVersion() int { return 1 }

// VsphereProviderSpec2 contains the fields of
// provider spec that the plugin expects
type VsphereProviderSpec2 struct {
	// Namespace is the vSphere workload namespace
	Namespace string `json:"namespace"`

	// ImageName describes the name of a VirtualMachineImage that is to be used as the base Operating System image of
	// the desired VirtualMachine instances.  The VirtualMachineImage resources can be introspected to discover identifying
	// attributes that may help users to identify the desired image to use.
	ImageName string `json:"imageName"`

	// ClassName describes the name of a VirtualMachineClass that is to be used as the overlaid resource configuration
	// of VirtualMachine.  A VirtualMachineClass is used to further customize the attributes of the VirtualMachine
	// instance.  See VirtualMachineClass for more description.
	ClassName string `json:"className"`

	// NetworkName is the name of the network for the virtualmachines.vmoperator.vmware.com resource
	NetworkName string `json:"networkName"`
	// NetworkType is the type of the network for the virtualmachines.vmoperator.vmware.com resource
	NetworkType string `json:"networkType"`

	// StorageClass describes the name of a StorageClass that should be used to configure storage-related attributes of the VirtualMachine
	// instance.
	// +optional
	StorageClass *string `json:"storageClass,omitempty"`

	// ResourcePolicyName describes the name of a VirtualMachineSetResourcePolicy to be used when creating the
	// VirtualMachine instance.
	// +optional
	ResourcePolicyName *string `json:"resourcePolicyName,omitempty"`

	// SystemDisk specifies the system disk
	// +optional
	SystemDisk *VSphereSystemDisk `json:"systemDisk,omitempty"`

	// TODO
	// ExtraConfig allows to specify additional VM options.
	// e.g. sched.swap.vmxSwapEnabled=false to disable the VMX process swap file
	// +optional
	//ExtraConfig map[string]string `json:"extraConfig,omitempty"`
}

// SpecVersion returns spec version
func (s *VsphereProviderSpec2) SpecVersion() int { return 2 }

// VSphereSystemDisk specifies system disk of a machine
type VSphereSystemDisk struct {
	// Size is disk size in GB
	Size int `json:"size"`
}

// VApp contains the properties of the VApp
type VApp struct {
	// Properties are the properties values of the VApp
	Properties map[string]string `json:"properties"`
}
