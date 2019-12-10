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
	"github.com/gardener/machine-controller-manager/pkg/apis/machine/v1alpha1"
	"github.com/vmware/govmomi"
)

type pseudoFlagKey string
type flagKey string

var (
	clientPseudoFlagKey = pseudoFlagKey("client")
	specPseudoFlagKey   = pseudoFlagKey("spec")
)

func ContextWithPseudoFlagset(ctx context.Context, client *govmomi.Client, spec *v1alpha1.VsphereMachineClassSpec) context.Context {
	ctx = context.WithValue(ctx, clientPseudoFlagKey, client)
	ctx = context.WithValue(ctx, specPseudoFlagKey, spec)
	return ctx
}

func GetClientFromPseudoFlagset(ctx context.Context) *govmomi.Client {
	return ctx.Value(clientPseudoFlagKey).(*govmomi.Client)
}

func GetSpecFromPseudoFlagset(ctx context.Context) *v1alpha1.VsphereMachineClassSpec {
	return ctx.Value(specPseudoFlagKey).(*v1alpha1.VsphereMachineClassSpec)
}
