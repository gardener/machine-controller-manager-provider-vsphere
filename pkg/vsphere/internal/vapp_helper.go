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

package vsphere

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

type coreosConfig struct {
	PasswdHash     string
	Hostname       string
	SSHKeys        []string
	UserdataBase64 string
}

func coreosIgnition(config *coreosConfig) (string, error) {
	ignitionTemplate := `{
  "ignition": {"config":{},"timeouts":{},"version":"2.1.0"},
  "networkd":{"units":[{"contents":"[Match]\nName=ens192\n\n[Network]\nDHCP=yes\nLinkLocalAddressing=no\nIPv6AcceptRA=no\n","name":"00-ens192.network"}]},
  "passwd":{"users":[{"name":"core","passwordHash":"{{.PasswdHash}}","sshAuthorizedKeys":[{{range $index,$elem := .SSHKeys}}{{if $index}},{{end}}"{{$elem}}"{{end}}]}]},
  "storage": {
	"directories":[{"filesystem":"root","path":"/var/lib/coreos-install"}],
	"files":[
	  {"filesystem":"root","path":"/etc/hostname","contents":{"source":"data:,{{.Hostname}}"},"mode":420},
	  {"filesystem":"root","path":"/var/lib/coreos-install/user_data","contents":{"source":"data:text/plain;charset=utf-8;base64,{{.UserdataBase64}}"},"mode":420}
	]
  },
  "systemd":{}
}
`
	tmpl, err := template.New("ignition").Parse(ignitionTemplate)
	if err != nil {
		return "", errors.Wrap(err, "Creating ignition file for CoreOS failed")
	}
	buf := bytes.NewBufferString("")
	err = tmpl.Execute(buf, config)
	if err != nil {
		return "", errors.Wrap(err, "Creating ignition file for CoreOS failed on executing template")
	}
	return buf.String(), nil
}

func addSshKeysSection(userdata string, sshKeys []string) (string, error) {
	if len(sshKeys) == 0 {
		return userdata, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(userdata)
	if err != nil {
		return "", errors.Wrap(err, "Decoding userdata failed")
	}
	s := string(decoded)
	if strings.Contains(s, "ssh_authorized_keys:") {
		return "", fmt.Errorf("userdata already contains key `ssh_authorized_keys`")
	}
	s = s + "\nssh_authorized_keys:\n"
	for _, key := range sshKeys {
		s = s + fmt.Sprintf("- %q\n", key)
	}
	return base64.StdEncoding.EncodeToString([]byte(s)), nil
}
