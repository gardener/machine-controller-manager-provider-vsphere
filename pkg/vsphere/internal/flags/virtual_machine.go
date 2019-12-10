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

package flags

import (
	"context"
	"github.com/vmware/govmomi/object"
)

type VirtualMachineFlag struct {
	*ClientFlag
	*DatacenterFlag
	*SearchFlag

	name string
	vm   *object.VirtualMachine
}

var virtualMachineFlagKey = flagKey("virtualMachine")

func NewVirtualMachineFlag(ctx context.Context) (*VirtualMachineFlag, context.Context) {
	if v := ctx.Value(virtualMachineFlagKey); v != nil {
		return v.(*VirtualMachineFlag), ctx
	}

	v := &VirtualMachineFlag{}
	v.name = GetSpecFromPseudoFlagset(ctx).TemplateVM
	v.ClientFlag, ctx = NewClientFlag(ctx)
	v.DatacenterFlag, ctx = NewDatacenterFlag(ctx)
	v.SearchFlag, ctx = NewSearchFlag(ctx, SearchVirtualMachines)
	ctx = context.WithValue(ctx, virtualMachineFlagKey, v)
	return v, ctx
}

func (flag *VirtualMachineFlag) VirtualMachine() (*object.VirtualMachine, error) {
	ctx := context.TODO()

	if flag.vm != nil {
		return flag.vm, nil
	}

	// Use search flags if specified.
	if flag.SearchFlag.IsSet() {
		vm, err := flag.SearchFlag.VirtualMachine()
		if err != nil {
			return nil, err
		}

		flag.vm = vm
		return flag.vm, nil
	}

	// Never look for a default virtual machine.
	if flag.name == "" {
		return nil, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	flag.vm, err = finder.VirtualMachine(ctx, flag.name)
	return flag.vm, err
}
