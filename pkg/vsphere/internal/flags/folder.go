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

type FolderFlag struct {
	*DatacenterFlag

	name   string
	folder *object.Folder
}

var folderFlagKey = flagKey("folder")

func NewFolderFlag(ctx context.Context) (*FolderFlag, context.Context) {
	if v := ctx.Value(folderFlagKey); v != nil {
		return v.(*FolderFlag), ctx
	}

	v := &FolderFlag{}
	v.name = GetSpecFromPseudoFlagset(ctx).Folder
	v.DatacenterFlag, ctx = NewDatacenterFlag(ctx)
	ctx = context.WithValue(ctx, folderFlagKey, v)
	return v, ctx
}

func (flag *FolderFlag) Folder() (*object.Folder, error) {
	if flag.folder != nil {
		return flag.folder, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	if flag.folder, err = finder.FolderOrDefault(context.TODO(), flag.name); err != nil {
		return nil, err
	}

	return flag.folder, nil
}

func (flag *FolderFlag) FolderOrDefault(kind string) (*object.Folder, error) {
	if flag.folder != nil {
		return flag.folder, nil
	}

	if flag.name != "" {
		return flag.Folder()
	}

	// RootFolder, no dc required
	if kind == "/" {
		client, err := flag.Client()
		if err != nil {
			return nil, err
		}

		flag.folder = object.NewRootFolder(client)
		return flag.folder, nil
	}

	dc, err := flag.Datacenter()
	if err != nil {
		return nil, err
	}

	folders, err := dc.Folders(context.TODO())
	if err != nil {
		return nil, err
	}

	switch kind {
	case "vm":
		flag.folder = folders.VmFolder
	case "host":
		flag.folder = folders.HostFolder
	case "datastore":
		flag.folder = folders.DatastoreFolder
	case "network":
		flag.folder = folders.NetworkFolder
	default:
		panic(kind)
	}

	return flag.folder, nil
}
