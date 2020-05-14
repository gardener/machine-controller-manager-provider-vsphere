/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

This file was copied and modified from the kubernetes-csi/drivers project
https://github.com/kubernetes-csi/drivers/blob/release-1.0/pkg/nfs/nodeserver.go

Modifications Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.
*/

package vsphere

import (
	"fmt"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	"golang.org/x/net/context"
	"k8s.io/klog"
)

// CreateMachine handles a machine creation request
// REQUIRED METHOD
//
// REQUEST PARAMETERS (driver.CreateMachineRequest)
// MachineName           string             Contains the name of the machine object for whom an VM is to be created at the provider
// ProviderSpec          bytes(blob)        Template/Configuration of the machine to be created is given by at the provider
// Secrets               map<string,bytes>  (Optional) Contains a map from string to string contains any cloud specific secrets that can be used by the provider
// LastKnownState        bytes(blob)        (Optional) Last known state of VM during last operation. Could be helpful to continue operation from previous state
//
// RESPONSE PARAMETERS (driver.CreateMachineResponse)
// ProviderID            string             Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//                                          ProviderID typically matches with the node.Spec.ProviderID on the node object.
//                                          Eg: gce://project-name/region/vm-ProviderID
// NodeName              string             Returns the name of the node-object that the VM register's with Kubernetes.
//                                          This could be different from req.MachineName as well
// LastKnownState        bytes(blob)        (Optional) Last known state of VM during the current operation.
//                                          Could be helpful to continue operations in future requests.
//
// OPTIONAL IMPLEMENTATION LOGIC
// It is optionally expected by the safety controller to use an identification mechanisms to map the VM Created by a providerSpec.
// These could be done using tag(s)/resource-groups etc.
// This logic is used by safety controller to delete orphan VMs which are not backed by any machine CRD
//
func (ms *MachinePlugin) CreateMachine(ctx context.Context, req *driver.CreateMachineRequest) (*driver.CreateMachineResponse, error) {
	// Log messages to track start of request
	klog.V(2).Infof("Create machine request has been received for %q", req.Machine.Name)

	providerSpec, err := decodeProviderSpecAndSecret(req.MachineClass, req.Secret)
	if err != nil {
		return nil, prepareErrorf(err, "Create machine %q failed on decodeProviderSpecAndSecret", req.Machine.Name)
	}

	providerID, err := ms.SPI.CreateMachine(ctx, req.Machine.Name, providerSpec, req.Secret)
	if err != nil {
		return nil, prepareErrorf(err, "Create machine %q failed", req.Machine.Name)
	}

	response := &driver.CreateMachineResponse{
		ProviderID:     providerID,
		NodeName:       req.Machine.Name,
		LastKnownState: fmt.Sprintf("Created %s", providerID),
	}

	klog.V(2).Infof("VM with Provider-ID: %q created for Machine: %q", response.ProviderID, req.Machine.Name)
	return response, nil
}

// DeleteMachine handles a machine deletion request
//
// REQUEST PARAMETERS (driver.DeleteMachineRequest)
// MachineName          string              Contains the name of the machine object for the backing VM(s) have to be deleted
// ProviderID           string              Contains the unique identification of the VM at the cloud provider
// ProviderSpec         bytes(blob)         Template/Configuration of the machine to be deleted is given by at the provider
// Secrets              map<string,bytes>   (Optional) Contains a map from string to string contains any cloud specific secrets that can be used by the provider
// LastKnownState       bytes(blob)         (Optional) Last known state of VM during last operation. Could be helpful to continue operation from previous state
//
// RESPONSE PARAMETERS (driver.DeleteMachineResponse)
// LastKnownState       bytes(blob)        (Optional) Last known state of VM during the current operation.
//                                          Could be helpful to continue operations in future requests.
//
func (ms *MachinePlugin) DeleteMachine(ctx context.Context, req *driver.DeleteMachineRequest) (*driver.DeleteMachineResponse, error) {
	// Log messages to track delete request
	klog.V(2).Infof("Machine deletion request has been received for %q", req.Machine.Name)

	providerSpec, err := decodeProviderSpecAndSecret(req.MachineClass, req.Secret)
	if err != nil {
		return nil, prepareErrorf(err, "Delete machine %q failed on decodeProviderSpecAndSecret", req.Machine.Name)
	}

	providerID, err := ms.SPI.DeleteMachine(ctx, req.Machine.Name, req.Machine.Spec.ProviderID, providerSpec, req.Secret)
	if err != nil {
		return nil, prepareErrorf(err, "Delete machine %q failed", req.Machine.Name)
	}

	klog.V(2).Infof("VM %q for Machine %q was terminated succesfully", providerID, req.Machine.Name)

	return &driver.DeleteMachineResponse{}, nil
}

// GetMachineStatus handles a machine get status request
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.GetMachineStatusRequest)
// MachineName          string              Contains the name of the machine object for whose status is to be retrived
// ProviderID           string              Contains the unique identification of the VM at the cloud provider
// ProviderSpec         bytes(blob)         Template/Configuration of the machine whose status is to be retrived
// Secrets              map<string,bytes>   (Optional) Contains a map from string to string contains any cloud specific secrets that can be used by the provider
//
// RESPONSE PARAMETERS (driver.GetMachineStatueResponse)
// ProviderID           string              Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//                                          ProviderID typically matches with the node.Spec.ProviderID on the node object.
//                                          Eg: gce://project-name/region/vm-ProviderID
// NodeName             string              Returns the name of the node-object that the VM register's with Kubernetes.
//                                          This could be different from req.MachineName as well
//
func (ms *MachinePlugin) GetMachineStatus(ctx context.Context, req *driver.GetMachineStatusRequest) (*driver.GetMachineStatusResponse, error) {
	// Log messages to track start of request
	klog.V(2).Infof("Machine status request has been received for %q", req.Machine.Name)

	providerSpec, err := decodeProviderSpecAndSecret(req.MachineClass, req.Secret)
	if err != nil {
		return nil, prepareErrorf(err, "Machine status %q failed on decodeProviderSpecAndSecret", req.Machine.Name)
	}

	providerID, err := ms.SPI.GetMachineStatus(ctx, req.Machine.Name, req.Machine.Spec.ProviderID, providerSpec, req.Secret)
	if err != nil {
		return nil, prepareErrorf(err, "Machine status %q failed", req.Machine.Name)
	}

	response := &driver.GetMachineStatusResponse{
		ProviderID: providerID,
		NodeName:   req.Machine.Name,
	}

	klog.V(2).Infof("Machine status: found VM %q for Machine: %q", response.ProviderID, req.Machine.Name)

	return response, nil
}

// ListMachines lists all the machines possibilly created by a providerSpec
// Identifying machines created by a given providerSpec depends on the OPTIONAL IMPLEMENTATION LOGIC
// you have used to identify machines created by a providerSpec. It could be tags/resource-groups etc
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (driver.ListMachinesRequest)
// ProviderSpec          bytes(blob)         Template/Configuration of the machine that wouldn've been created by this ProviderSpec (Machine Class)
// Secrets               map<string,bytes>   (Optional) Contains a map from string to string contains any cloud specific secrets that can be used by the provider
//
// RESPONSE PARAMETERS (driver.ListMachinesResponse)
// MachineList           map<string,string>  A map containing the keys as the MachineID and value as the MachineName
//                                           for all machine's who where possibilly created by this ProviderSpec
//
func (ms *MachinePlugin) ListMachines(ctx context.Context, req *driver.ListMachinesRequest) (*driver.ListMachinesResponse, error) {
	// Log messages to track start of request
	klog.V(2).Infof("List machines request has been received")

	providerSpec, err := decodeProviderSpecAndSecret(req.MachineClass, req.Secret)
	if err != nil {
		return nil, prepareErrorf(err, "List machines failed on decodeProviderSpecAndSecret")
	}

	machineList, err := ms.SPI.ListMachines(ctx, providerSpec, req.Secret)
	if err != nil {
		return nil, prepareErrorf(err, "List machines failed")
	}

	klog.V(2).Infof("List machines request for dc %s, folder %s found %d machines", providerSpec.Datacenter, providerSpec.Folder, len(machineList))
	return &driver.ListMachinesResponse{
		MachineList: machineList,
	}, nil
}

// GetVolumeIDs returns a list of Volume IDs for all PV Specs for whom an provider volume was found
//
// REQUEST PARAMETERS (driver.GetVolumeIDsRequest)
// PVSpecList            bytes(blob)         PVSpecsList is a list PV specs for whom volume-IDs are required. Plugin should parse this raw data into pre-defined list of PVSpecs.
//
// RESPONSE PARAMETERS (driver.GetVolumeIDsResponse)
// VolumeIDs             repeated string     VolumeIDs is a repeated list of VolumeIDs.
//
func (ms *MachinePlugin) GetVolumeIDs(ctx context.Context, req *driver.GetVolumeIDsRequest) (*driver.GetVolumeIDsResponse, error) {
	// Log messages to track start of request
	klog.V(2).Infof("GetVolumeIDs request has been received")
	klog.V(4).Infof("PVSpecList = %q", req.PVSpecs)

	var volumeIDs []string
	for i := range req.PVSpecs {
		spec := req.PVSpecs[i]
		if spec.VsphereVolume == nil {
			// Not an vsphere volume
			continue
		}
		volumeID := spec.VsphereVolume.VolumePath
		volumeIDs = append(volumeIDs, volumeID)
	}

	klog.V(2).Infof("GetVolumeIDs machines request has been processed successfully (%d/%d).", len(volumeIDs), len(req.PVSpecs))
	klog.V(4).Infof("GetVolumeIDs volumneIDs: %v", volumeIDs)

	Resp := &driver.GetVolumeIDsResponse{
		VolumeIDs: volumeIDs,
	}
	return Resp, nil
}
