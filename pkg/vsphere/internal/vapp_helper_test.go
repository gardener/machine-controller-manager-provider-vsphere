/*
 *
 *  * Copyright 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
 *  *
 *  * Licensed under the Apache License, Version 2.0 (the "License");
 *  * you may not use this file except in compliance with the License.
 *  * You may obtain a copy of the License at
 *  *
 *  *      http://www.apache.org/licenses/LICENSE-2.0
 *  *
 *  * Unless required by applicable law or agreed to in writing, software
 *  * distributed under the License is distributed on an "AS IS" BASIS,
 *  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  * See the License for the specific language governing permissions and
 *  *
 *
 */

package vsphere

import (
	"github.com/onsi/gomega"
	"testing"
)

const expectedContent = `{
  "ignition": {"config":{},"timeouts":{},"version":"2.1.0"},
  "networkd":{"units":[{"contents":"[Match]\nName=ens192\n\n[Network]\nDHCP=yes\nLinkLocalAddressing=no\nIPv6AcceptRA=no\n","name":"00-ens192.network"}]},
  "passwd":{"users":[{"name":"core","passwordHash":"$1$9H6.uffe$e5XfhfWO4EcT8JdUvzEOT0","sshAuthorizedKeys":["ssh1","ssh2"]}]},
  "storage": {
	"directories":[{"filesystem":"root","path":"/var/lib/coreos-install"}],
	"files":[
	  {"filesystem":"root","path":"/etc/hostname","contents":{"source":"data:,foo"},"mode":420},
	  {"filesystem":"root","path":"/var/lib/coreos-install/user_data","contents":{"source":"data:text/plain;charset=utf-8;base64,XYZ"},"mode":420}
	]
  },
  "systemd":{}
}
`

func TestCoreOSIgnition(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	config := &coreosConfig{
		PasswdHash:     "$1$9H6.uffe$e5XfhfWO4EcT8JdUvzEOT0",
		Hostname:       "foo",
		SSHKeys:        []string{"ssh1", "ssh2"},
		UserdataBase64: "XYZ",
	}
	content, err := coreosIgnition(config)
	if err != nil {
		t.Errorf("coreosIgnition failed with %s", err)
	}

	g.Expect(content).To(gomega.Equal(expectedContent))
}

func TestAddSSHKeys(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	userdata := "I2Nsb3VkLWNvbmZpZwpydW5jbWQ6Ci0gJ2VjaG8gMTI3LjAuMC4xIGBob3N0bmFtZWAgPj4gL2V0Yy9ob3N0cy14eHgnCg=="

	newUserdata, err := addSshKeysSection(userdata, []string{"ssh1", "ssh2"})
	if err != nil {
		t.Errorf("addSshKeysSection failed with %s", err)
	}
	g.Expect(newUserdata).To(gomega.Equal("I2Nsb3VkLWNvbmZpZwpydW5jbWQ6Ci0gJ2VjaG8gMTI3LjAuMC4xIGBob3N0bmFtZWAgPj4gL2V0Yy9ob3N0cy14eHgnCgpzc2hfYXV0aG9yaXplZF9rZXlzOgotICJzc2gxIgotICJzc2gyIgo="))
}
