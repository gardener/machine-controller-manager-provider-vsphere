/*
 * Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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
 * limitations under the License.
 *
 */

package tags

import (
	"testing"

	"github.com/onsi/gomega"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
)

const (
	cluster1 = "cluster"
	cluster2 = "cluster2"
	role     = "node"
)

var (
	oldSpecTags = map[string]string{
		api.TagClusterPrefix + cluster1: "1",
		api.TagNodeRolePrefix + role:    "1",
	}

	oldSpecTags2 = map[string]string{
		api.TagClusterPrefix + cluster2: "1",
		api.TagNodeRolePrefix + role:    "1",
	}

	newSpecTags = map[string]string{
		api.TagMCMClusterName: cluster1,
		api.TagMCMRole:        role,
	}

	newSpecTags2 = map[string]string{
		api.TagMCMClusterName: cluster2,
		api.TagMCMRole:        role,
	}
)

func TestOldTags(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	tags, errs := NewRelevantTags(oldSpecTags)
	g.Expect(errs).To(gomega.BeNil())

	g.Expect(tags.Matches(oldSpecTags)).To(gomega.Equal(true))
	g.Expect(tags.Matches(oldSpecTags2)).To(gomega.Equal(false))
	g.Expect(tags.Matches(newSpecTags)).To(gomega.Equal(true))
	g.Expect(tags.Matches(newSpecTags2)).To(gomega.Equal(false))
}

func TestNewTags(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	tags, errs := NewRelevantTags(newSpecTags)
	g.Expect(errs).To(gomega.BeNil())

	g.Expect(tags.Matches(oldSpecTags)).To(gomega.Equal(true))
	g.Expect(tags.Matches(oldSpecTags2)).To(gomega.Equal(false))
	g.Expect(tags.Matches(newSpecTags)).To(gomega.Equal(true))
	g.Expect(tags.Matches(newSpecTags2)).To(gomega.Equal(false))
}

func TestEmptyTags(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	tags, errs := NewRelevantTags(map[string]string{})
	g.Expect(tags).To(gomega.BeNil())
	g.Expect(len(errs)).To(gomega.Equal(2))
}
