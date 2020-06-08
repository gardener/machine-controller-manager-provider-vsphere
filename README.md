# machine-controller-manager-provider-vsphere

This project contains the external/out-of-tree plugin (driver) implementation of machine-controller-manager for
VMware `vSphere`.

## Prerequisites

- vSphere installation, user with permissions to create/delete VMs in the specified data center. 
  Details see below [Recommended permissions for role of vSphere user](#Recommended permissions for role of vSphere user)
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
    for more details) but this is still in early stage.

## Supported vSphere versions

This provider has been tested with vSphere 6.7 (Update 3) and vSphere 7.0

## Recommended permissions for role of vSphere user 

The vSphere user used for this provider should have been assigned to a role including these permissions
(use vSphere Client / Menu Administration / Access Control / Role to define a role and assign it to the user
with `Global Permissions`)

* Datastore 
  * Allocate space 
  * Browse datastore 
  * Low level file operations 
  * Remove file 
  * Update virtual machine files 
  * Update virtual machine metadata 
* Global 
  * Cancel task 
  * Manage custom attributes 
  * Set custom attribute 
* Network 
  * Assign network 
* Resource 
  * Assign virtual machine to resource pool 
* Tasks 
  * Create task 
  * Update task 
* vApp 
  * Add virtual machine 
  * Assign resource pool 
  * Assign vApp 
  * Clone 
  * Power off 
  * Power on 
  * View OVF environment 
  * vApp application configuration 
  * vApp instance configuration 
  * vApp managedBy configuration 
  * vApp resource configuration 
* Virtual machine 
  * Change Configuration 
    * Acquire disk lease 
    * Add existing disk 
    * Add new disk 
    * Add or remove device 
    * Advanced configuration 
    * Change CPU count 
    * Change Memory 
    * Change Settings 
    * Change Swapfile placement 
    * Change resource 
    * Configure Host USB device 
    * Configure Raw device 
    * Configure managedBy 
    * Display connection settings 
    * Extend virtual disk 
    * Modify device settings 
    * Query Fault Tolerance compatibility 
    * Query unowned files 
    * Reload from path 
    * Remove disk 
    * Rename 
    * Reset guest information 
    * Set annotation 
    * Toggle disk change tracking 
    * Toggle fork parent 
    * Upgrade virtual machine compatibility 
  * Edit Inventory 
    * Create from existing 
    * Create new 
    * Move 
    * Register 
    * Remove 
    * Unregister 
  * Guest operations 
    * Guest operation alias modification 
    * Guest operation alias query 
    * Guest operation modifications 
    * Guest operation program execution 
    * Guest operation queries 
  * Interaction 
    * Power off 
    * Power on 
    * Reset 
  * Provisioning 
    * Allow disk access 
    * Allow file access 
    * Allow read-only disk access 
    * Allow virtual machine files upload 
    * Clone template 
    * Clone virtual machine 
    * Customize guest 
    * Deploy template 
    * Mark as virtual machine 
    * Modify customization specification 
    * Promote disks 
    * Read customization specifications
