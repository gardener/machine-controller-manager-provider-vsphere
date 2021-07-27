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
	"fmt"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

type DatacenterFlag struct {
	*ClientFlag

	Name   string
	dc     *object.Datacenter
	finder *find.Finder
	err    error
}

var datacenterFlagKey = flagKey("datacenter")

func NewDatacenterFlag(ctx context.Context) (*DatacenterFlag, context.Context) {
	if v := ctx.Value(datacenterFlagKey); v != nil {
		return v.(*DatacenterFlag), ctx
	}

	v := &DatacenterFlag{}
	v.Name = GetSpecFromPseudoFlagset(ctx).Datacenter
	v.ClientFlag, ctx = NewClientFlag(ctx)
	ctx = context.WithValue(ctx, datacenterFlagKey, v)
	return v, ctx
}

func (flag *DatacenterFlag) Finder(all ...bool) (*find.Finder, error) {
	if flag.finder != nil {
		return flag.finder, nil
	}

	c, err := flag.Client()
	if err != nil {
		return nil, err
	}

	allFlag := false
	if len(all) == 1 {
		allFlag = all[0]
	}
	finder := find.NewFinder(c, allFlag)

	// Datacenter is not required (ls command for example).
	// Set for relative func if dc flag is given or
	// if there is a single (default) Datacenter
	ctx := context.TODO()
	if flag.Name == "" {
		flag.dc, flag.err = finder.DefaultDatacenter(ctx)
	} else {
		if flag.dc, err = finder.Datacenter(ctx, flag.Name); err != nil {
			return nil, err
		}
	}

	finder.SetDatacenter(flag.dc)

	flag.finder = finder

	return flag.finder, nil
}

func (flag *DatacenterFlag) Datacenter() (*object.Datacenter, error) {
	if flag.dc != nil {
		return flag.dc, nil
	}

	_, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	if flag.err != nil {
		// Should only happen if no dc is specified and len(dcs) > 1
		return nil, flag.err
	}

	return flag.dc, err
}

func (flag *DatacenterFlag) DatacenterIfSpecified() (*object.Datacenter, error) {
	if flag.Name == "" {
		return nil, nil
	}
	return flag.Datacenter()
}

func (flag *DatacenterFlag) ManagedObject(ctx context.Context, arg string) (types.ManagedObjectReference, error) {
	var ref types.ManagedObjectReference

	if ref.FromString(arg) {
		return ref, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return ref, err
	}

	l, err := finder.ManagedObjectList(ctx, arg)
	if err != nil {
		return ref, err
	}

	switch len(l) {
	case 0:
		return ref, fmt.Errorf("%s not found", arg)
	case 1:
		return l[0].Object.Reference(), nil
	default:
		var objs []types.ManagedObjectReference
		for _, o := range l {
			objs = append(objs, o.Object.Reference())
		}
		return ref, fmt.Errorf("%d objects at path %q: %s", len(l), arg, objs)
	}
}

func (flag *DatacenterFlag) ManagedObjects(ctx context.Context, args []string) ([]types.ManagedObjectReference, error) {
	var refs []types.ManagedObjectReference

	c, err := flag.Client()
	if err != nil {
		return nil, err
	}

	if len(args) == 0 {
		refs = append(refs, c.ServiceContent.RootFolder)
		return refs, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	for _, arg := range args {
		var ref types.ManagedObjectReference
		if ref.FromString(arg) {
			// e.g. output from object.collect
			refs = append(refs, ref)
			continue
		}
		elements, err := finder.ManagedObjectList(ctx, arg)
		if err != nil {
			return nil, err
		}

		if len(elements) == 0 {
			return nil, fmt.Errorf("object '%s' not found", arg)
		}

		for _, e := range elements {
			refs = append(refs, e.Object.Reference())
		}
	}

	return refs, nil
}
