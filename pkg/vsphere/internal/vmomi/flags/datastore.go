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
	"github.com/vmware/govmomi/vim25/types"
)

type DatastoreFlag struct {
	*DatacenterFlag

	Name string

	ds *object.Datastore
}

var datastoreFlagKey = flagKey("datastore")

// NewCustomDatastoreFlag creates and returns a new DatastoreFlag without
// trying to retrieve an existing one from the specified context.
func NewCustomDatastoreFlag(ctx context.Context) (*DatastoreFlag, context.Context) {
	v := &DatastoreFlag{}
	v.DatacenterFlag, ctx = NewDatacenterFlag(ctx)
	return v, ctx
}

func NewDatastoreFlag(ctx context.Context) (*DatastoreFlag, context.Context) {
	if v := ctx.Value(datastoreFlagKey); v != nil {
		return v.(*DatastoreFlag), ctx
	}

	v, ctx := NewCustomDatastoreFlag(ctx)
	v.Name = GetSpecFromPseudoFlagset(ctx).Datastore
	ctx = context.WithValue(ctx, datastoreFlagKey, v)
	return v, ctx
}

func (f *DatastoreFlag) Args(args []string) []object.DatastorePath {
	var files []object.DatastorePath

	for _, arg := range args {
		var p object.DatastorePath

		if p.FromString(arg) {
			f.Name = p.Datastore
		} else {
			p.Datastore = f.Name
			p.Path = arg
		}

		files = append(files, p)
	}

	return files
}

func (f *DatastoreFlag) Datastore() (*object.Datastore, error) {
	if f.ds != nil {
		return f.ds, nil
	}

	var p object.DatastorePath
	if p.FromString(f.Name) {
		// Example use case:
		//   -ds "$(govc object.collect -s vm/foo config.files.logDirectory)"
		f.Name = p.Datastore
	}

	finder, err := f.Finder()
	if err != nil {
		return nil, err
	}

	if f.ds, err = finder.DatastoreOrDefault(context.TODO(), f.Name); err != nil {
		return nil, err
	}

	return f.ds, nil
}

func (flag *DatastoreFlag) DatastoreIfSpecified() (*object.Datastore, error) {
	if flag.Name == "" {
		return nil, nil
	}
	return flag.Datastore()
}

func (f *DatastoreFlag) DatastorePath(name string) (string, error) {
	ds, err := f.Datastore()
	if err != nil {
		return "", err
	}

	return ds.Path(name), nil
}

func (f *DatastoreFlag) Stat(ctx context.Context, file string) (types.BaseFileInfo, error) {
	ds, err := f.Datastore()
	if err != nil {
		return nil, err
	}

	return ds.Stat(ctx, file)

}
