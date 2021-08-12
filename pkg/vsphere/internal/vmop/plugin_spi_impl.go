/*
 * Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
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

package vmop

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis/tags"
	"github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/errors"
	vmopapi "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/klog/v2"
	ctrlClient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	api "github.com/gardener/machine-controller-manager-provider-vsphere/pkg/vsphere/apis"
)

// PluginSPIImpl is the real implementation of PluginSPI interface
// that makes the calls to the provider SDK
type PluginSPIImpl struct{}

const providerPrefix = "vsphere://"

const (
	timeout     = 3 * time.Minute
	createCheck = 10 * time.Second
)

var scheme *runtime.Scheme
var clientCache *cache.LRUExpireCache

func init() {
	scheme = runtime.NewScheme()
	vmopapi.AddToScheme(scheme)
	corev1.AddToScheme(scheme)

	clientCache = cache.NewLRUExpireCache(10)
}

// CreateMachine creates a VM by cloning from a template
func (spi *PluginSPIImpl) CreateMachine(ctx context.Context, machineName string, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (string, error) {
	client, err := createVsphereKubernetesClient(ctx, secrets)
	if err != nil {
		return "", fmt.Errorf("creating vsphere k8s client failed: %w", err)
	}

	userdata, err := addSSHKeysSection(string(secrets.Data["userData"]), providerSpec.SSHKeys)
	if err != nil {
		return "", fmt.Errorf("adding ssh keys to userdata failed: %w", err)
	}

	v2 := providerSpec.V2
	if v2 == nil {
		return "", fmt.Errorf("missing v2")
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName(machineName),
			Namespace: v2.Namespace,
		},
	}
	_, err = controllerutil.CreateOrUpdate(ctx, client, configMap, func() error {
		configMap.Data = map[string]string{
			"hostname":  machineName,
			"user-data": base64.StdEncoding.EncodeToString([]byte(userdata)),
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("creating/updating configmap %s failed: %w", configMap.Name, err)
	}

	relevantTags, _ := tags.NewRelevantTags(providerSpec.Tags)
	if relevantTags == nil {
		return "", fmt.Errorf("missing relevant tags")
	}

	vm := createEmptyVirtualMachine(machineName, v2.Namespace)
	vm.Spec.ClassName = v2.ClassName
	vm.Spec.NetworkInterfaces = []vmopapi.VirtualMachineNetworkInterface{
		{NetworkType: v2.NetworkType, NetworkName: v2.NetworkName},
	}
	if v2.StorageClass != nil {
		vm.Spec.StorageClass = *v2.StorageClass
	}
	if v2.ResourcePolicyName != nil {
		vm.Spec.ResourcePolicyName = *v2.ResourcePolicyName
	}
	vm.Spec.ImageName = v2.ImageName
	vm.Spec.VmMetadata = &vmopapi.VirtualMachineMetadata{
		ConfigMapName: configMap.Name,
		Transport:     vmopapi.VirtualMachineMetadataOvfEnvTransport,
	}
	vm.Annotations = relevantTags.NonRelevant(providerSpec.Tags)
	vm.Annotations["vmoperator.vmware.com/image-supported-check"] = "disable"
	vm.Annotations["vmoperator.vmware.com/vsphere-customization"] = "disable"
	vm.Labels = relevantTags.GetLabels()
	vm.Labels[api.LabelMCMVSphere] = "true"
	vm.Spec.PowerState = vmopapi.VirtualMachinePoweredOn

	if v2.SystemDisk != nil {
		deviceKey := 2000 // 2000 is a de facto value that is assigned to the first disk when a VM is created.
		value := fmt.Sprintf("%dGi", v2.SystemDisk.Size)
		q, err := resource.ParseQuantity(value)
		if err != nil {
			return "", fmt.Errorf("cannot parse disk size quantity %s: %s", value, err)
		}
		vm.Spec.Volumes = []vmopapi.VirtualMachineVolume{
			{
				Name: "system",
				VsphereVolume: &vmopapi.VsphereVolumeSource{
					Capacity:  corev1.ResourceList{"ephemeral-storage": q},
					DeviceKey: &deviceKey,
				},
			},
		}
	}

	err = client.Create(ctx, vm)
	if err != nil {
		return "", fmt.Errorf("creating virtual machine %s failed: %w", machineName, err)
	}

	timeoutTimestamp := time.Now().Add(timeout)
	for time.Now().Before(timeoutTimestamp) && vm.Status.InstanceUUID == "" {
		time.Sleep(createCheck)
		if err := client.Get(ctx, objectKeyFromObject(vm), vm); err != nil {
			return "", fmt.Errorf("getting virtual machine %s failed: %w", machineName, err)
		}
	}
	if vm.Status.InstanceUUID == "" {
		_ = client.Delete(ctx, vm)
		return "", fmt.Errorf("timeout on vm create of virtual machine %s. phase=%s", machineName, vm.Status.Phase)
	}

	providerID := spi.encodeProviderID(v2.Namespace, vm.Status.InstanceUUID)
	return providerID, nil
}

func (spi *PluginSPIImpl) encodeProviderID(namespace, machineID string) string {
	if machineID == "" {
		return ""
	}
	return fmt.Sprintf("%s%s/%s", providerPrefix, namespace, machineID)
}

func (spi *PluginSPIImpl) decodeProviderID(providerID string) (namespace, machineID string) {
	if !strings.HasPrefix(providerID, providerPrefix) {
		return
	}
	parts := strings.Split(providerID[len(providerPrefix):], "/")
	if len(parts) != 2 {
		return
	}
	namespace = parts[0]
	machineID = parts[1]
	return
}

// DeleteMachine deletes a VM by name
func (spi *PluginSPIImpl) DeleteMachine(ctx context.Context, machineName string, providerID string, providerSpec *api.VsphereProviderSpec2, secrets *corev1.Secret) (string, error) {
	client, err := createVsphereKubernetesClient(ctx, secrets)
	if err != nil {
		return "", fmt.Errorf("creating vsphere k8s client failed: %w", err)
	}

	vm := createEmptyVirtualMachine(machineName, providerSpec.Namespace)
	if err := client.Get(ctx, objectKeyFromObject(vm), vm); err != nil {
		if apierrors.IsNotFound(err) {
			return "", &errors.MachineNotFoundError{Name: machineName, Namespace: providerSpec.Namespace}
		}
		return "", fmt.Errorf("getting virtual machine %s failed: %w", machineName, err)
	}
	if err := client.Delete(ctx, vm); err != nil {
		return "", fmt.Errorf("deleting virtual machine %s failed: %w", machineName, err)
	}
	foundProviderID := spi.encodeProviderID(providerSpec.Namespace, vm.Status.InstanceUUID)
	return foundProviderID, nil
}

// ShutDownMachine shuts down a machine by name
func (spi *PluginSPIImpl) ShutDownMachine(ctx context.Context, machineName string, providerID string, providerSpec *api.VsphereProviderSpec2, secrets *corev1.Secret) (string, error) {
	client, err := createVsphereKubernetesClient(ctx, secrets)
	if err != nil {
		return "", fmt.Errorf("creating vsphere k8s client failed: %w", err)
	}

	vm := createEmptyVirtualMachine(machineName, providerSpec.Namespace)
	if err := client.Get(ctx, objectKeyFromObject(vm), vm); err != nil {
		if apierrors.IsNotFound(err) {
			return "", &errors.MachineNotFoundError{Name: machineName, Namespace: providerSpec.Namespace}
		}
		return "", fmt.Errorf("getting virtual machine %s failed: %w", machineName, err)
	}
	vm.Spec.PowerState = vmopapi.VirtualMachinePoweredOff
	if err := client.Update(ctx, vm); err != nil {
		return "", fmt.Errorf("updating virtual machine %s failed: %w", machineName, err)
	}
	foundProviderID := spi.encodeProviderID(providerSpec.Namespace, vm.Status.InstanceUUID)
	return foundProviderID, nil
}

// GetMachineStatus checks for existence of VM by name
func (spi *PluginSPIImpl) GetMachineStatus(ctx context.Context, machineName string, providerID string, providerSpec *api.VsphereProviderSpec2, secrets *corev1.Secret) (string, error) {
	client, err := createVsphereKubernetesClient(ctx, secrets)
	if err != nil {
		return "", fmt.Errorf("creating vsphere k8s client failed: %w", err)
	}

	vm := createEmptyVirtualMachine(machineName, providerSpec.Namespace)
	if err := client.Get(ctx, objectKeyFromObject(vm), vm); err != nil {
		if apierrors.IsNotFound(err) {
			return "", &errors.MachineNotFoundError{Name: machineName, Namespace: providerSpec.Namespace}
		}
		return "", fmt.Errorf("getting virtual machine %s failed: %w", machineName, err)
	}
	klog.V(4).Infof("Machine %s has status: phase=%s, power=%s, instanceUUID=%s, uniqueID=%s, vmIP=%s",
		machineName, vm.Status.Phase, vm.Status.PowerState, vm.Status.InstanceUUID, vm.Status.UniqueID, vm.Status.VmIp)
	foundProviderID := spi.encodeProviderID(providerSpec.Namespace, vm.Status.InstanceUUID)
	return foundProviderID, nil
}

// ListMachines lists all VMs in the DC or folder
func (spi *PluginSPIImpl) ListMachines(ctx context.Context, providerSpec *api.VsphereProviderSpec, secrets *corev1.Secret) (map[string]string, error) {
	client, err := createVsphereKubernetesClient(ctx, secrets)
	if err != nil {
		return nil, fmt.Errorf("creating vsphere k8s client failed: %w", err)
	}

	v2 := providerSpec.V2
	if v2 == nil {
		return nil, fmt.Errorf("missing v2")
	}

	machineList := map[string]string{}
	relevantTags, _ := tags.NewRelevantTags(providerSpec.Tags)
	if relevantTags == nil {
		return machineList, nil
	}

	vms := &vmopapi.VirtualMachineList{}
	labels := relevantTags.GetLabels()
	labels[api.LabelMCMVSphere] = "true"
	err = client.List(ctx, vms, ctrlClient.InNamespace(v2.Namespace), ctrlClient.MatchingLabels(labels))
	if err != nil {
		return nil, fmt.Errorf("listing virtual machines in namespace %s failed: %w", v2.Namespace, err)
	}

	for _, vm := range vms.Items {
		machineName := vm.Name
		if vm.Status.InstanceUUID != "" {
			providerID := spi.encodeProviderID(v2.Namespace, vm.Status.InstanceUUID)
			machineList[providerID] = machineName
		}
	}

	klog.V(2).Infof("List machines request for namespace %s found %d machines", v2.Namespace, len(machineList))

	return machineList, nil
}

func hashMD5(in []byte) string {
	return fmt.Sprintf("%x", md5.Sum(in))
}

func createVsphereKubernetesClient(ctx context.Context, secret *corev1.Secret) (ctrlClient.Client, error) {
	kubeconfig, ok := secret.Data[api.VSphereKubeconfig]
	if !ok {
		return nil, fmt.Errorf("missing %s key in secret", api.VSphereKubeconfig)
	}
	hash := hashMD5(kubeconfig)
	if value, ok := clientCache.Get(hash); ok {
		return value.(ctrlClient.Client), nil
	}

	client, err := createRealVsphereKubernetesClient(ctx, kubeconfig)
	if err != nil {
		return nil, err
	}

	clientCache.Add(hash, client, 1*time.Hour)
	return client, nil
}

func createRealVsphereKubernetesClient(ctx context.Context, kubeconfig []byte) (ctrlClient.Client, error) {
	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*clientcmdapi.Config, error) {
		return clientcmd.Load([]byte(kubeconfig))
	})
	if err != nil {
		return nil, fmt.Errorf("build config from kubeconfig failed: %w", err)
	}

	if config.QPS == 0 {
		config.QPS = 5
	}
	if config.Burst == 0 {
		config.Burst = 10
	}

	client, err := ctrlClient.New(
		rest.AddUserAgent(config, "machine-controller-manager-provider-vsphere"),
		ctrlClient.Options{
			Scheme: scheme,
		})

	return client, err
}

func createEmptyVirtualMachine(machineName, namespace string) *vmopapi.VirtualMachine {
	return &vmopapi.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      machineName,
			Namespace: namespace,
		},
	}
}

func configMapName(machineName string) string {
	return fmt.Sprintf("vm-metadata-%s", machineName)
}

func addSSHKeysSection(userdata string, sshKeys []string) (string, error) {
	if len(sshKeys) == 0 {
		return userdata, nil
	}
	s := userdata
	if strings.Contains(s, "ssh_authorized_keys:") {
		return "", fmt.Errorf("userdata already contains key `ssh_authorized_keys`")
	}
	s = s + "\nssh_authorized_keys:\n"
	for _, key := range sshKeys {
		s = s + fmt.Sprintf("- %q\n", key)
	}
	return s, nil
}

func objectKeyFromObject(obj metav1.Object) ctrlClient.ObjectKey {
	return ctrlClient.ObjectKey{Namespace: obj.GetNamespace(), Name: obj.GetName()}
}
