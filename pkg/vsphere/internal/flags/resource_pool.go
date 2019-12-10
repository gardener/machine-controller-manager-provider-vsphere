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

type ResourcePoolFlag struct {
	*DatacenterFlag

	name string
	pool *object.ResourcePool
}

var resourcePoolFlagKey = flagKey("resourcePool")

func NewResourcePoolFlag(ctx context.Context) (*ResourcePoolFlag, context.Context) {
	if v := ctx.Value(resourcePoolFlagKey); v != nil {
		return v.(*ResourcePoolFlag), ctx
	}

	v := &ResourcePoolFlag{}
	v.name = GetSpecFromPseudoFlagset(ctx).Pool
	v.DatacenterFlag, ctx = NewDatacenterFlag(ctx)
	ctx = context.WithValue(ctx, resourcePoolFlagKey, v)
	return v, ctx
}

func (flag *ResourcePoolFlag) ResourcePool() (*object.ResourcePool, error) {
	if flag.pool != nil {
		return flag.pool, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	if flag.pool, err = finder.ResourcePoolOrDefault(context.TODO(), flag.name); err != nil {
		return nil, err
	}

	return flag.pool, nil
}

func (flag *ResourcePoolFlag) ResourcePoolIfSpecified() (*object.ResourcePool, error) {
	if flag.name == "" {
		return nil, nil
	}
	return flag.ResourcePool()
}
