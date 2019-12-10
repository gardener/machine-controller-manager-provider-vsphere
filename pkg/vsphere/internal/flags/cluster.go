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
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type ClusterFlag struct {
	*DatacenterFlag

	Name string

	cluster *object.ClusterComputeResource
	pc      *property.Collector
}

var clusterFlagKey = flagKey("cluster")

func NewClusterFlag(ctx context.Context) (*ClusterFlag, context.Context) {
	if v := ctx.Value(clusterFlagKey); v != nil {
		return v.(*ClusterFlag), ctx
	}

	v := &ClusterFlag{}
	v.Name = GetSpecFromPseudoFlagset(ctx).ComputeCluster
	v.DatacenterFlag, ctx = NewDatacenterFlag(ctx)
	ctx = context.WithValue(ctx, clusterFlagKey, v)
	return v, ctx
}

func (f *ClusterFlag) Cluster() (*object.ClusterComputeResource, error) {
	if f.cluster != nil {
		return f.cluster, nil
	}

	finder, err := f.Finder()
	if err != nil {
		return nil, err
	}

	if f.cluster, err = finder.ClusterComputeResourceOrDefault(context.TODO(), f.Name); err != nil {
		return nil, err
	}

	f.pc = property.DefaultCollector(f.cluster.Client())

	return f.cluster, nil
}

func (f *ClusterFlag) ClusterIfSpecified() (*object.ClusterComputeResource, error) {
	if f.Name == "" {
		return nil, nil
	}
	return f.Cluster()
}

func (f *ClusterFlag) objectMap(ctx context.Context, kind string, names []string) (map[string]types.ManagedObjectReference, error) {
	cluster, err := f.Cluster()
	if err != nil {
		return nil, err
	}

	objects := make(map[string]types.ManagedObjectReference, len(names))
	for _, name := range names {
		objects[name] = types.ManagedObjectReference{}
	}

	m := view.NewManager(cluster.Client())
	v, err := m.CreateContainerView(ctx, cluster.Reference(), []string{kind}, true)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = v.Destroy(ctx)
	}()

	var entities []mo.ManagedEntity

	err = v.Retrieve(ctx, []string{"ManagedEntity"}, []string{"name"}, &entities)
	if err != nil {
		return nil, err
	}

	for _, e := range entities {
		if _, ok := objects[e.Name]; ok {
			objects[e.Name] = e.Self
		}
	}

	for name, ref := range objects {
		if ref.Value == "" {
			return nil, fmt.Errorf("%s %q not found", kind, name)
		}
	}

	return objects, nil
}

func (f *ClusterFlag) ObjectList(ctx context.Context, kind string, names []string) ([]types.ManagedObjectReference, error) {
	objs, err := f.objectMap(ctx, kind, names)
	if err != nil {
		return nil, err
	}

	var refs []types.ManagedObjectReference

	for _, name := range names { // preserve order
		refs = append(refs, objs[name])
	}

	return refs, nil
}

func (f *ClusterFlag) Names(ctx context.Context, refs []types.ManagedObjectReference) (map[types.ManagedObjectReference]string, error) {
	names := make(map[types.ManagedObjectReference]string, len(refs))

	if len(refs) != 0 {
		var objs []mo.ManagedEntity
		err := f.pc.Retrieve(ctx, refs, []string{"name"}, &objs)
		if err != nil {
			return nil, err
		}

		for _, obj := range objs {
			names[obj.Self] = obj.Name
		}
	}

	return names, nil
}
