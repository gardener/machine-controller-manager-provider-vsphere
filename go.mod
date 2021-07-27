module github.com/gardener/machine-controller-manager-provider-vsphere

go 1.15

require (
	github.com/gardener/machine-controller-manager v0.39.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	github.com/vmware-tanzu/vm-operator-api v0.1.4-0.20210722184632-99fee0b6197e
	github.com/vmware/govmomi v0.22.1
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/net v0.0.0-20210428140749-89ef3d95e781
	k8s.io/api v0.17.11
	k8s.io/apimachinery v0.17.11
	k8s.io/client-go v0.17.11
	k8s.io/component-base v0.17.11
	k8s.io/klog/v2 v2.9.0
	sigs.k8s.io/controller-runtime v0.5.10
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2 // keep this value in sync with sigs.k8s.io/controller-runtime	k8s.io/api => k8s.io/api v0.17.11
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.11
	k8s.io/apiserver => k8s.io/apiserver v0.17.11
	k8s.io/client-go => k8s.io/client-go v0.17.11
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.11
	k8s.io/kube-openapi => github.com/gardener/kube-openapi v0.0.0-20200807191151-9232ec702af2
)
