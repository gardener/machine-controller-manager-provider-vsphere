/*
 * Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package vsphere

import (
	"context"

	"github.com/gardener/machine-controller-manager/pkg/util/provider/driver"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/fake"
)

const (
	// FailAtMethodNotImplemented is the error returned for methods which are not yet implemented
	FailAtMethodNotImplemented string = "rpc error: code = Unimplemented desc = "
)

var _ = Describe("#MachineController", func() {
	vspherePVSpec := &corev1.PersistentVolumeSpec{
		PersistentVolumeSource: corev1.PersistentVolumeSource{
			VsphereVolume: &corev1.VsphereVirtualDiskVolumeSource{
				VolumePath: "dummyPD",
			},
		},
	}
	vspherePVCSISpec := &corev1.PersistentVolumeSpec{
		PersistentVolumeSource: corev1.PersistentVolumeSource{
			CSI: &corev1.CSIPersistentVolumeSource{
				Driver:       "csi.vsphere.vmware.com",
				VolumeHandle: "csiPD",
			},
		},
	}
	vspherePVSpecEmptyPD := &corev1.PersistentVolumeSpec{
		PersistentVolumeSource: corev1.PersistentVolumeSource{},
	}
	vspherePVResponse := []string{"dummyPD", "csiPD"}

	var ms *MachinePlugin
	var fakePluginSPIImpl *fake.PluginSPIImpl

	var _ = BeforeSuite(func() {
		fakePluginSPIImpl = &fake.PluginSPIImpl{}
		ms = &MachinePlugin{
			SPI: fakePluginSPIImpl,
		}
	})
	Describe("##GetVolumeIDs", func() {
		type setup struct {
		}
		type action struct {
			machineRequest *driver.GetVolumeIDsRequest
		}
		type expect struct {
			machineResponse   *driver.GetVolumeIDsResponse
			errToHaveOccurred bool
			errMessage        string
		}
		type data struct {
			setup  setup
			action action
			expect expect
		}

		DescribeTable("###table",
			func(data *data) {

				ctx := context.Background()
				resp, err := ms.GetVolumeIDs(ctx, data.action.machineRequest)
				if data.expect.errToHaveOccurred {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(data.expect.errMessage))
				} else {
					Expect(err).ToNot(HaveOccurred())
					Expect(len(resp.VolumeIDs)).To(Equal(len(data.expect.machineResponse.VolumeIDs)))
					for i, r := range resp.VolumeIDs {
						Expect(r).To(Equal(data.expect.machineResponse.VolumeIDs[i]))
					}
				}

			},

			Entry("With valid PV list", &data{
				action: action{
					machineRequest: &driver.GetVolumeIDsRequest{
						PVSpecs: []*corev1.PersistentVolumeSpec{vspherePVSpec, vspherePVCSISpec},
					},
				},
				expect: expect{
					machineResponse: &driver.GetVolumeIDsResponse{
						VolumeIDs: vspherePVResponse,
					},
					errToHaveOccurred: false,
					errMessage:        FailAtMethodNotImplemented,
				},
			}),
			Entry("With emtpy PV list", &data{
				action: action{
					machineRequest: &driver.GetVolumeIDsRequest{
						PVSpecs: []*corev1.PersistentVolumeSpec{vspherePVSpecEmptyPD},
					},
				},
				expect: expect{
					machineResponse: &driver.GetVolumeIDsResponse{
						VolumeIDs: []string{},
					},
					errToHaveOccurred: false,
				},
			}),
		)
	})

})
