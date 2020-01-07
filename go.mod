module github.com/gardener/machine-controller-manager-provider-vsphere

go 1.13

require (
	github.com/gardener/machine-spec v0.5.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.7.1
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/vmware/govmomi v0.21.1-0.20190909001527-8d286461ab92
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	golang.org/x/sys v0.0.0-20190712062909-fae7ac547cb7 // indirect
	google.golang.org/genproto v0.0.0-20190716160619-c506a9f90610 // indirect
	google.golang.org/grpc v1.22.0
	gopkg.in/inf.v0 v0.9.1 // indirect
	k8s.io/api v0.0.0-20191114100352-16d7abae0d2a // kubernetes-1.16.3
	sigs.k8s.io/yaml v1.1.0
)
