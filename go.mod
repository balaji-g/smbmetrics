module github.com/samba-in-kubernetes/smbmetrics

go 1.16

require (
	github.com/go-logr/logr v0.4.0
	github.com/prometheus/client_golang v1.11.1
	github.com/shirou/gopsutil/v3 v3.24.5
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.9.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.1
)
