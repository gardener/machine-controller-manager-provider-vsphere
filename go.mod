module github.com/gardener/machine-controller-manager-provider-vsphere

go 1.16

require (
	github.com/gardener/machine-controller-manager v0.42.0
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/vmware/govmomi v0.22.1
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/net v0.0.0-20210326060303-6b1517762897
	k8s.io/api v0.20.6
	k8s.io/component-base v0.20.6
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	k8s.io/api => k8s.io/api v0.20.6 // v0.20.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.6 // v0.20.6
	k8s.io/apiserver => k8s.io/apiserver v0.20.6 // v0.20.6
	k8s.io/client-go => k8s.io/client-go v0.20.6 // v0.20.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.6 // v0.20.6
)
