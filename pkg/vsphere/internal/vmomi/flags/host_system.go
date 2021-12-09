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

type HostSystemFlag struct {
	*ClientFlag
	*DatacenterFlag
	*SearchFlag

	name string
	host *object.HostSystem
	pool *object.ResourcePool
}

var hostSystemFlagKey = flagKey("hostSystem")

func NewHostSystemFlag(ctx context.Context) (*HostSystemFlag, context.Context) {
	if v := ctx.Value(hostSystemFlagKey); v != nil {
		return v.(*HostSystemFlag), ctx
	}

	v := &HostSystemFlag{}
	v.name = GetSpecFromPseudoFlagset(ctx).HostSystem
	v.ClientFlag, ctx = NewClientFlag(ctx)
	v.DatacenterFlag, ctx = NewDatacenterFlag(ctx)
	v.SearchFlag, ctx = NewSearchFlag(ctx, SearchHosts)
	ctx = context.WithValue(ctx, hostSystemFlagKey, v)
	return v, ctx
}

func (flag *HostSystemFlag) HostSystemIfSpecified() (*object.HostSystem, error) {
	if flag.host != nil {
		return flag.host, nil
	}

	// Use search flags if specified.
	if flag.SearchFlag.IsSet() {
		host, err := flag.SearchFlag.HostSystem()
		if err != nil {
			return nil, err
		}

		flag.host = host
		return flag.host, nil
	}

	// Never look for a default host system.
	// A host system parameter is optional for vm creation. It uses a mandatory
	// resource pool parameter to determine where the vm should be placed.
	if flag.name == "" {
		return nil, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	flag.host, err = finder.HostSystem(context.TODO(), flag.name)
	return flag.host, err
}

func (flag *HostSystemFlag) HostSystem() (*object.HostSystem, error) {
	host, err := flag.HostSystemIfSpecified()
	if err != nil {
		return nil, err
	}

	if host != nil {
		return host, nil
	}

	finder, err := flag.Finder()
	if err != nil {
		return nil, err
	}

	flag.host, err = finder.DefaultHostSystem(context.TODO())
	return flag.host, err
}

func (flag *HostSystemFlag) HostNetworkSystem() (*object.HostNetworkSystem, error) {
	host, err := flag.HostSystem()
	if err != nil {
		return nil, err
	}

	return host.ConfigManager().NetworkSystem(context.TODO())
}
