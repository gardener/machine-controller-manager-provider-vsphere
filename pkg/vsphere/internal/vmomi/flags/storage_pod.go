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

type StoragePodFlag struct {
	*DatacenterFlag

	Name string

	sp *object.StoragePod
}

var storagePodFlagKey = flagKey("storagePod")

func NewStoragePodFlag(ctx context.Context) (*StoragePodFlag, context.Context) {
	if v := ctx.Value(storagePodFlagKey); v != nil {
		return v.(*StoragePodFlag), ctx
	}

	v := &StoragePodFlag{}
	v.Name = GetSpecFromPseudoFlagset(ctx).DatastoreCluster
	v.DatacenterFlag, ctx = NewDatacenterFlag(ctx)
	ctx = context.WithValue(ctx, storagePodFlagKey, v)
	return v, ctx
}

func (f *StoragePodFlag) Isset() bool {
	return f.Name != ""
}

func (f *StoragePodFlag) StoragePod() (*object.StoragePod, error) {
	ctx := context.TODO()
	if f.sp != nil {
		return f.sp, nil
	}

	finder, err := f.Finder()
	if err != nil {
		return nil, err
	}

	if f.Isset() {
		f.sp, err = finder.DatastoreCluster(ctx, f.Name)
		if err != nil {
			return nil, err
		}
	} else {
		f.sp, err = finder.DefaultDatastoreCluster(ctx)
		if err != nil {
			return nil, err
		}
	}

	return f.sp, nil
}
