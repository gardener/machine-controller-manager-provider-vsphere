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
	"encoding/json"
	"fmt"

	"github.com/gardener/machine-spec/lib/go/cmi"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
)

// NOTE
//
// The basic working of the controller will work with just implementing the CreateMachine() & DeleteMachine() methods.
// You can first implement these two methods and check the working of the controller.
// Once this works you can implement the rest of the methods.

// CreateMachine handles a machine creation request
// REQUIRED METHOD
//
// REQUEST PARAMETERS (cmi.CreateMachineRequest)
// MachineName           string             Contains the name of the machine object for whom an VM is to be created at the provider
// ProviderSpec          bytes(blob)        Template/Configuration of the machine to be created is given by at the provider
// Secrets               map<string,bytes>  (Optional) Contains a map from string to string contains any cloud specific secrets that can be used by the provider
// LastKnownState        bytes(blob)        (Optional) Last known state of VM during last operation. Could be helpful to continue operation from previous state
//
// RESPONSE PARAMETERS (cmi.CreateMachineResponse)
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
func (ms *MachinePlugin) CreateMachine(ctx context.Context, req *cmi.CreateMachineRequest) (*cmi.CreateMachineResponse, error) {
	// Log messages to track start of request
	glog.V(2).Infof("Create machine request has been recieved for %q", req.MachineName)

	providerSpec, secrets, err := decodeProviderSpecAndSecret(req.ProviderSpec, req.Secrets, true)
	if err != nil {
		return nil, prepareInternalErrorf(err, "Create machine %q failed on decodeProviderSpecAndSecret", req.MachineName)
	}

	machineID, err := ms.SPI.CreateMachine(ctx, req.MachineName, providerSpec, secrets)
	if err != nil {
		return nil, prepareInternalErrorf(err, "Create machine %q failed", req.MachineName)
	}

	response := &cmi.CreateMachineResponse{
		ProviderID:     encodeProviderID(machineID),
		NodeName:       req.MachineName,
		LastKnownState: []byte(fmt.Sprintf("Created %s", machineID)),
	}

	glog.V(2).Infof("VM with Provider-ID: %q created for Machine: %q", response.ProviderID, req.MachineName)
	return response, nil
}

// DeleteMachine handles a machine deletion request
//
// REQUEST PARAMETERS (cmi.DeleteMachineRequest)
// MachineName          string              Contains the name of the machine object for the backing VM(s) have to be deleted
// ProviderSpec         bytes(blob)         Template/Configuration of the machine to be deleted is given by at the provider
// Secrets              map<string,bytes>   (Optional) Contains a map from string to string contains any cloud specific secrets that can be used by the provider
// LastKnownState       bytes(blob)         (Optional) Last known state of VM during last operation. Could be helpful to continue operation from previous state
//
// RESPONSE PARAMETERS (cmi.DeleteMachineResponse)
// LastKnownState       bytes(blob)        (Optional) Last known state of VM during the current operation.
//                                          Could be helpful to continue operations in future requests.
//
func (ms *MachinePlugin) DeleteMachine(ctx context.Context, req *cmi.DeleteMachineRequest) (*cmi.DeleteMachineResponse, error) {
	// Log messages to track delete request
	glog.V(2).Infof("Machine deletion request has been received for %q", req.MachineName)

	providerSpec, secrets, err := decodeProviderSpecAndSecret(req.ProviderSpec, req.Secrets, false)
	if err != nil {
		return nil, prepareInternalErrorf(err, "Delete machine %q failed on decodeProviderSpecAndSecret", req.MachineName)
	}

	machineID, err := ms.SPI.DeleteMachine(ctx, req.MachineName, providerSpec, secrets)
	if err != nil {
		return nil, prepareInternalErrorf(err, "Delete machine %q failed", req.MachineName)
	}

	glog.V(2).Infof("VM %q for Machine %q was terminated succesfully", machineID, req.MachineName)

	return &cmi.DeleteMachineResponse{}, nil
}

// GetMachineStatus handles a machine get status request
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (cmi.GetMachineStatusRequest)
// MachineName          string              Contains the name of the machine object for whose status is to be retrived
// ProviderSpec         bytes(blob)         Template/Configuration of the machine whose status is to be retrived
// Secrets              map<string,bytes>   (Optional) Contains a map from string to string contains any cloud specific secrets that can be used by the provider
//
// RESPONSE PARAMETERS (cmi.GetMachineStatueResponse)
// ProviderID           string              Unique identification of the VM at the cloud provider. This could be the same/different from req.MachineName.
//                                          ProviderID typically matches with the node.Spec.ProviderID on the node object.
//                                          Eg: gce://project-name/region/vm-ProviderID
// NodeName             string              Returns the name of the node-object that the VM register's with Kubernetes.
//                                          This could be different from req.MachineName as well
//
func (ms *MachinePlugin) GetMachineStatus(ctx context.Context, req *cmi.GetMachineStatusRequest) (*cmi.GetMachineStatusResponse, error) {
	// Log messages to track start of request
	glog.V(2).Infof("Machine status request has been received for %q", req.MachineName)

	providerSpec, secrets, err := decodeProviderSpecAndSecret(req.ProviderSpec, req.Secrets, false)
	if err != nil {
		return nil, prepareInternalErrorf(err, "Machine status %q failed on decodeProviderSpecAndSecret", req.MachineName)
	}

	machineID, err := ms.SPI.GetMachineStatus(ctx, req.MachineName, providerSpec, secrets)
	if err != nil {
		return nil, prepareInternalErrorf(err, "Machine status %q failed", req.MachineName)
	}

	return &cmi.GetMachineStatusResponse{
		ProviderID: encodeProviderID(machineID),
		NodeName:   req.MachineName,
	}, nil
}

// ListMachines lists all the machines possibilly created by a providerSpec
// Identifying machines created by a given providerSpec depends on the OPTIONAL IMPLEMENTATION LOGIC
// you have used to identify machines created by a providerSpec. It could be tags/resource-groups etc
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (cmi.ListMachinesRequest)
// ProviderSpec          bytes(blob)         Template/Configuration of the machine that wouldn've been created by this ProviderSpec (Machine Class)
// Secrets               map<string,bytes>   (Optional) Contains a map from string to string contains any cloud specific secrets that can be used by the provider
//
// RESPONSE PARAMETERS (cmi.ListMachinesResponse)
// MachineList           map<string,string>  A map containing the keys as the MachineID and value as the MachineName
//                                           for all machine's who where possibilly created by this ProviderSpec
//
func (ms *MachinePlugin) ListMachines(ctx context.Context, req *cmi.ListMachinesRequest) (*cmi.ListMachinesResponse, error) {
	// Log messages to track start of request
	glog.V(2).Infof("List machines request has been received for %q", req.ProviderSpec)

	providerSpec, secrets, err := decodeProviderSpecAndSecret(req.ProviderSpec, req.Secrets, false)
	if err != nil {
		return nil, prepareInternalErrorf(err, "List machines failed on decodeProviderSpecAndSecret")
	}

	machineIDList, err := ms.SPI.ListMachines(ctx, providerSpec, secrets)
	if err != nil {
		return nil, prepareInternalErrorf(err, "List machines failed")
	}

	machineList := map[string]string{}

	for id, name := range machineIDList {
		machineList[encodeProviderID(id)] = name
	}

	return &cmi.ListMachinesResponse{
		MachineList: machineList,
	}, nil
}

// GetVolumeIDs returns a list of Volume IDs for all PV Specs for whom an provider volume was found
//
// REQUEST PARAMETERS (cmi.GetVolumeIDsRequest)
// PVSpecList            bytes(blob)         PVSpecsList is a list PV specs for whom volume-IDs are required. Plugin should parse this raw data into pre-defined list of PVSpecs.
//
// RESPONSE PARAMETERS (cmi.GetVolumeIDsResponse)
// VolumeIDs             repeated string     VolumeIDs is a repeated list of VolumeIDs.
//
func (ms *MachinePlugin) GetVolumeIDs(ctx context.Context, req *cmi.GetVolumeIDsRequest) (*cmi.GetVolumeIDsResponse, error) {
	// Log messages to track start of request
	glog.V(2).Infof("GetVolumeIDs request has been received for %q", req.PVSpecList)

	var (
		volumeIDs   []string
		volumeSpecs []*corev1.PersistentVolumeSpec
	)

	// Log messages to track start and end of request
	glog.V(2).Infof("GetVolumeIDs request has been recieved for %q", req.PVSpecList)

	err := json.Unmarshal(req.PVSpecList, &volumeSpecs)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	for i := range volumeSpecs {
		spec := volumeSpecs[i]
		if spec.VsphereVolume == nil {
			// Not an aws volume
			continue
		}
		volumeID := spec.VsphereVolume.VolumePath
		volumeIDs = append(volumeIDs, volumeID)
	}

	glog.V(2).Infof("GetVolumeIDs machines request has been processed successfully. \nList: %v", volumeIDs)

	Resp := &cmi.GetVolumeIDsResponse{
		VolumeIDs: volumeIDs,
	}
	return Resp, nil
}

// ShutDownMachine handles a machine shutdown/power-off/stop request
// OPTIONAL METHOD
//
// REQUEST PARAMETERS (cmi.ShutDownMachineRequest)
// ProviderSpec          bytes(blob)         Template/Configuration of the machine that wouldn've been created by this ProviderSpec (Machine Class)
// Secrets               map<string,bytes>   (Optional) Contains a map from string to string contains any cloud specific secrets that can be used by the provider
// LastKnownState        bytes(blob)        (Optional) Last known state of VM during last operation. Could be helpful to continue operation from previous state
//
// RESPONSE PARAMETERS (cmi.DeleteMachineResponse)
// LastKnownState        bytes(blob)        (Optional) Last known state of VM during the current operation.
//                                          Could be helpful to continue operations in future requests.
//
func (ms *MachinePlugin) ShutDownMachine(ctx context.Context, req *cmi.ShutDownMachineRequest) (*cmi.ShutDownMachineResponse, error) {
	// Log messages to track start of request
	glog.V(2).Infof("ShutDown machine request has been received for %q", req.MachineName)

	providerSpec, secrets, err := decodeProviderSpecAndSecret(req.ProviderSpec, req.Secrets, false)
	if err != nil {
		return nil, prepareInternalErrorf(err, "ShutDown machine %q failed on decodeProviderSpecAndSecret", req.MachineName)
		return nil, err
	}

	machineID, err := ms.SPI.ShutDownMachine(ctx, req.MachineName, providerSpec, secrets)
	if err != nil {
		return nil, prepareInternalErrorf(err, "ShutDown machine %q failed", req.MachineName)
	}

	glog.V(2).Infof("VM %q for Machine %q was shutted down succesfully", machineID, req.MachineName)

	return &cmi.ShutDownMachineResponse{LastKnownState: []byte(fmt.Sprintf("Shutted down %s", machineID))}, nil
}
