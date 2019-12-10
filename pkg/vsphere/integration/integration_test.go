/*
 * Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package integration

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/internal"
	"sigs.k8s.io/yaml"
)

type integrationConfig struct {
	MachineName  string                   `json:"machineName"`
	ProviderSpec *api.VsphereProviderSpec `json:"providerSpec"`
	Secrets      *api.Secrets             `json:"secrets"`
}

// TestPluginSPIImpl tests creation and deleting of a VM via vSphere API.
// Path to configuration needs to be specified as environment variable MCM_PROVIDER_VSPHERE_CONFIG.
func TestPluginSPIImpl(t *testing.T) {
	configPath := os.Getenv("MCM_PROVIDER_VSPHERE_CONFIG")
	if configPath == "" {
		t.Skipf("No path to integrationConfig specified by environmental variable MCM_PROVIDER_VSPHERE_CONFIG")
		return
	}

	content, err := ioutil.ReadFile(configPath)
	if err != nil {
		t.Errorf("reading integrationConfig from %s failed with %s", configPath, err)
		return
	}

	cfg := integrationConfig{}
	err = yaml.Unmarshal([]byte(content), &cfg)
	if err != nil {
		t.Errorf("Unmarshalling integrationConfig failed with %s", err)
		return
	}

	spi := &internal.PluginSPIImpl{}
	ctx := context.TODO()
	machineID, err := spi.CreateMachine(ctx, cfg.MachineName, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("CreateMachine failed with %s", err)
		return
	}

	machineID2, err := spi.GetMachineStatus(ctx, cfg.MachineName, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("GetMachineStatus failed with %s", err)
		return
	}
	if machineID != machineID2 {
		t.Errorf("MachineID mismatch %s != %s", machineID, machineID2)
	}

	machineIDList, err := spi.ListMachines(ctx, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("ListMachines failed with %s", err)
	}

	found := false
	for id, name := range machineIDList {
		if id == machineID {
			if name != cfg.MachineName {
				t.Errorf("MachineName mismatch %s != %s", machineID, machineID2)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("Created machine with ID %s not found", machineID)
	}

	machineID2, err = spi.ShutDownMachine(ctx, cfg.MachineName, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("ShutDownMachine failed with %s", err)
	}
	if machineID != machineID2 {
		t.Errorf("MachineID mismatch %s != %s", machineID, machineID2)
	}

	machineID2, err = spi.DeleteMachine(ctx, cfg.MachineName, cfg.ProviderSpec, cfg.Secrets)
	if err != nil {
		t.Errorf("DeleteMachine failed with %s", err)
	}
	if machineID != machineID2 {
		t.Errorf("MachineID mismatch %s != %s", machineID, machineID2)
	}
}
