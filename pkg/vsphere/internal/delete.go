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

package internal

import (
	"context"
	"fmt"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
	errors2 "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/errors"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/internal/flags"
	"github.com/pkg/errors"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

func FindByIPath(ctx context.Context, client *govmomi.Client, spec *api.VsphereProviderSpec, machineName string) (*object.VirtualMachine, error) {
	ctx = flags.ContextWithPseudoFlagset(ctx, client, spec)
	searchFlag, ctx := flags.NewSearchFlag(ctx, flags.SearchVirtualMachines)

	folder := "vm"
	if spec.Folder != "" {
		folder = fmt.Sprintf("vm/%s", spec.Folder)
	}
	ipath := fmt.Sprintf("/%s/%s/%s", spec.Datacenter, folder, machineName)
	searchFlag.SetByInventoryPath(ipath)
	obj, err := searchFlag.VirtualMachine()
	if err != nil {
		switch err.(type) {
		case *flags.NotFoundError:
			return nil, &errors2.MachineNotFoundError{Name: machineName}
		default:
			return nil, errors.Wrapf(err, "find by inventory path %q failed", ipath)
		}
	}
	return obj, nil
}

type VirtualMachineVisitor func(uuid string, obj mo.ManagedEntity, field object.CustomFieldDefList) error

func VisitVirtualMachines(ctx context.Context, client *govmomi.Client, spec *api.VsphereProviderSpec, visitor VirtualMachineVisitor) error {
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
	ids := make([]string, 0, len(refs))
	morefs := make([]types.ManagedObjectReference, 0, len(refs))
	for _, ref := range refs {
		vm, ok := ref.(*object.VirtualMachine)
		if ok {
			ids = append(ids, vm.UUID(ctx))
			morefs = append(morefs, vm.Reference())
		}
	}

	var objs []mo.ManagedEntity
	err = property.DefaultCollector(client.Client).Retrieve(ctx, morefs, []string{"name", "customValue"}, &objs)
	if err != nil {
		return errors.Wrap(err, "DefaultCollector failed")
	}
	objMap := map[types.ManagedObjectReference]mo.ManagedEntity{}
	for _, obj := range objs {
		objMap[obj.Self] = obj
	}

	m, err := object.GetCustomFieldsManager(client.Client)
	if err != nil {
		return errors.Wrap(err, "GetCustomFieldsManager failed")
	}
	field, err := m.Field(ctx)
	if err != nil {
		return errors.Wrap(err, "Field failed")
	}

	for i := range ids {
		obj := objMap[morefs[i]]
		err := visitor(ids[i], obj, field)
		if err != nil {
			return errors.Wrapf(err, "visiting vm %s failed", obj.Name)
		}
	}

	return nil
}

func ShutDown(ctx context.Context, client *govmomi.Client, spec *api.VsphereProviderSpec, machineName string) (string, error) {
	vm, err := shutdown(ctx, client, spec, machineName)
	if err != nil {
		return "", err
	}
	return vm.UUID(ctx), nil
}

func Delete(ctx context.Context, client *govmomi.Client, spec *api.VsphereProviderSpec, machineName string) (string, error) {
	vm, err := shutdown(ctx, client, spec, machineName)
	if err != nil {
		return "", err
	}
	machineID := vm.UUID(ctx)

	task, err := vm.Destroy(ctx)
	if err != nil {
		return "", errors.Wrap(err, "starting Destroy failed")
	}
	_, err = task.WaitForResult(ctx, nil)
	if err != nil {
		return "", errors.Wrap(err, "Destroy failed")
	}
	return machineID, nil
}

func shutdown(ctx context.Context, client *govmomi.Client, spec *api.VsphereProviderSpec, machineName string) (*object.VirtualMachine, error) {
	vm, err := FindByIPath(ctx, client, spec, machineName)
	if err != nil {
		return nil, err
	}
	powerState, err := vm.PowerState(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "PowerState failed")
	}
	if powerState == types.VirtualMachinePowerStatePoweredOn {
		task, err := vm.PowerOff(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "starting PowerOff failed")
		}
		_, err = task.WaitForResult(ctx, nil)
		if err != nil {
			return nil, errors.Wrap(err, "PowerOff failed")
		}
	}
	return vm, nil
}
