/*
 * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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
 *
 */

package vsphere

import (
	"context"
	"fmt"
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/gardener/machine-controller-manager/pkg/driver/vsphere/flags"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func Find(ctx context.Context, client *govmomi.Client, spec *v1alpha1.VsphereMachineClassSpec, machineID string) (*object.VirtualMachine, error) {
	ctx = flags.ContextWithPseudoFlagset(ctx, client, spec)
	searchFlag, ctx := flags.NewSearchFlag(ctx, flags.SearchVirtualMachines)

	searchFlag.SetByUUID(machineID)
	return searchFlag.VirtualMachine()
}

type VirtualMachineVisitor func(machine *object.VirtualMachine, obj mo.ManagedEntity, field object.CustomFieldDefList) error

func VisitVirtualMachines(ctx context.Context, client *govmomi.Client, spec *v1alpha1.VsphereMachineClassSpec, visitor VirtualMachineVisitor) error {
	ctx = flags.ContextWithPseudoFlagset(ctx, client, spec)
	datacenterFlag, ctx := flags.NewDatacenterFlag(ctx)
	dc, err := datacenterFlag.Datacenter()
	if err != nil {
		return err
	}
	folders, err := dc.Folders(ctx)
	if err != nil {
		return err
	}
	folder := folders.VmFolder
	if spec.Folder != "" {
		refs, err := folder.Children(ctx)
		if err != nil {
			return err
		}
		folder = nil
		for _, ref := range refs {
			f, ok := ref.(*object.Folder)
			if ok {
				name, err := f.ObjectName(ctx)
				if err != nil {
					return err
				}
				if name == spec.Folder {
					folder = f
					break
				}
			}
		}
		if folder == nil {
			return fmt.Errorf("Folder %s not found", spec.Folder)
		}
	}

	refs, err := folder.Children(ctx)
	if err != nil {
		return err
	}
	vms := make([]*object.VirtualMachine, 0, len(refs))
	morefs := make([]types.ManagedObjectReference, 0, len(refs))
	for _, ref := range refs {
		vm, ok := ref.(*object.VirtualMachine)
		if ok {
			vms = append(vms, vm)
			morefs = append(morefs, vm.Reference())
		}
	}

	var objs []mo.ManagedEntity
	err = property.DefaultCollector(client.Client).Retrieve(ctx, morefs, []string{"name", "customValue"}, &objs)
	if err != nil {
		return errors.Wrap(err, "DefaultCollector failed")
	}
	m, err := object.GetCustomFieldsManager(client.Client)
	if err != nil {
		return errors.Wrap(err, "GetCustomFieldsManager failed")
	}
	field, err := m.Field(ctx)
	if err != nil {
		return errors.Wrap(err, "Field failed")
	}

	for i, vm := range vms {
		obj := objs[i]
		err := visitor(vm, obj, field)
		if err != nil {
			return errors.Wrapf(err, "visiting vm %s failed", obj.Name)
		}
	}

	return nil
}

func Delete(ctx context.Context, client *govmomi.Client, spec *v1alpha1.VsphereMachineClassSpec, machineID string) error {
	vm, err := Find(ctx, client, spec, machineID)
	if err != nil {
		return errors.Wrap(err, "find by machineID failed")
	}
	powerState, err := vm.PowerState(ctx)
	if err != nil {
		return errors.Wrap(err, "PowerState failed")
	}
	if powerState == types.VirtualMachinePowerStatePoweredOn {
		task, err := vm.PowerOff(ctx)
		if err != nil {
			return errors.Wrap(err, "starting PowerOff failed")
		}
		_, err = task.WaitForResult(ctx, nil)
		if err != nil {
			return errors.Wrap(err, "PowerOff failed")
		}
	}
	task, err := vm.Destroy(ctx)
	if err != nil {
		return errors.Wrap(err, "starting Destroy failed")
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "Destroy failed")
	}
	return nil
}
