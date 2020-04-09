module github.com/openfaas/faas-netes

go 1.13

require (
	github.com/google/go-cmp v0.3.0
	github.com/gorilla/context v1.1.1 // indirect
	github.com/gorilla/mux v1.6.2
	github.com/openfaas-incubator/openfaas-operator v0.0.0-20190923090851-8a2f0aa0fb60
	github.com/openfaas/faas v0.0.0-20191125105239-365f459b3f3a
	github.com/openfaas/faas-provider v0.0.0-20200101101649-8f7c35975e1b
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.2
	go.uber.org/goleak v1.0.0 // indirect
	k8s.io/api v0.17.4
	k8s.io/apimachinery v0.17.4
	k8s.io/client-go v0.17.4
	k8s.io/code-generator v0.17.1
	k8s.io/klog v1.0.0
)
