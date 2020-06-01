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
	"fmt"
	"strings"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
)

// RelevantTags contains tags keys and values for cluster name and role
type RelevantTags struct {
	clusterNameKey string
	clusterName    string
	nodeRoleKey    string
	nodeRole       string
}

// NewRelevantTags creates RelevantTags from given tag map.
// Returns nil if no complete set of tags provided.
func NewRelevantTags(tags map[string]string) (*RelevantTags, []error) {
	var clusterNameKey, clusterName, nodeRoleKey, nodeRole string
	for key, value := range tags {
		if strings.HasPrefix(key, api.TagClusterPrefix) {
			clusterNameKey = key
			clusterName = clusterNameKey[len(api.TagClusterPrefix):]
		} else if strings.HasPrefix(key, api.TagNodeRolePrefix) {
			nodeRoleKey = key
			nodeRole = nodeRoleKey[len(api.TagNodeRolePrefix):]
		} else if key == api.TagMCMClusterName {
			clusterNameKey = api.TagClusterPrefix + value
			clusterName = value
		} else if key == api.TagMCMRole {
			nodeRoleKey = api.TagNodeRolePrefix + value
			nodeRole = value
		}
	}

	var allErrs []error
	if clusterNameKey == "" {
		allErrs = append(allErrs, fmt.Errorf("tag required of the form '%s=****' or '%s****=1'", api.TagMCMClusterName, api.TagClusterPrefix))
	}
	if nodeRoleKey == "" {
		allErrs = append(allErrs, fmt.Errorf("tag required of the form '%s=****' or '%s****=1'", api.TagMCMRole, api.TagNodeRolePrefix))
	}
	if allErrs != nil {
		return nil, allErrs
	}

	return &RelevantTags{
		clusterNameKey: clusterNameKey,
		clusterName:    clusterName,
		nodeRoleKey:    nodeRoleKey,
		nodeRole:       nodeRole,
	}, nil
}

// Matches checks if given tags matches cluster name and role
func (t *RelevantTags) Matches(tags map[string]string) bool {
	matchedCluster := false
	matchedRole := false
	for key, value := range tags {
		switch key {
		case t.clusterNameKey:
			matchedCluster = true
		case t.nodeRoleKey:
			matchedRole = true
		case api.TagMCMClusterName:
			if value == t.clusterName {
				matchedCluster = true
			}
		case api.TagMCMRole:
			if value == t.nodeRole {
				matchedRole = true
			}
		}
	}
	return matchedCluster && matchedRole
}
