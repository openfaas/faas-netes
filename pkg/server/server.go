package server

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/openfaas/faas-netes/pkg/config"

	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	"github.com/openfaas/faas-netes/pkg/handlers"
	"github.com/openfaas/faas-netes/pkg/k8s"
	faasnetesk8s "github.com/openfaas/faas-netes/pkg/k8s"
	bootstrap "github.com/openfaas/faas-provider"
	v1apps "k8s.io/client-go/listers/apps/v1"

	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/proxy"
	"github.com/openfaas/faas-provider/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	coreinformer "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/kubernetes"
	glog "k8s.io/klog"
)

// TODO: Move to config pattern used else-where across project

const defaultHTTPPort = 8081
const defaultReadTimeout = 8
const defaultWriteTimeout = 8

// New creates HTTP server struct
func New(client clientset.Interface,
	kube kubernetes.Interface,
	endpointsInformer coreinformer.EndpointsInformer,
	deploymentLister v1apps.DeploymentLister,
	clusterRole bool,
	cfg config.BootstrapConfig) *Server {

	functionNamespace := "openfaas-fn"
	if namespace, exists := os.LookupEnv("function_namespace"); exists {
		functionNamespace = namespace
	}

	pprof := "false"
	if val, exists := os.LookupEnv("pprof"); exists {
		pprof = val
	}

	lister := endpointsInformer.Lister()
	functionLookup := k8s.NewFunctionLookup(functionNamespace, lister)

	bootstrapConfig := types.FaaSConfig{
		ReadTimeout:  cfg.FaaSConfig.ReadTimeout,
		WriteTimeout: cfg.FaaSConfig.WriteTimeout,
		TCPPort:      cfg.FaaSConfig.TCPPort,
		EnableHealth: true,
	}

	bootstrapHandlers := types.FaaSHandlers{
		FunctionProxy:        proxy.NewHandlerFunc(bootstrapConfig, functionLookup),
		DeleteHandler:        makeDeleteHandler(functionNamespace, client),
		DeployHandler:        makeApplyHandler(functionNamespace, client),
		FunctionReader:       makeListHandler(functionNamespace, client, deploymentLister),
		ReplicaReader:        makeReplicaReader(functionNamespace, client, deploymentLister),
		ReplicaUpdater:       makeReplicaHandler(functionNamespace, kube),
		UpdateHandler:        makeApplyHandler(functionNamespace, client),
		HealthHandler:        makeHealthHandler(),
		InfoHandler:          makeInfoHandler(),
		SecretHandler:        handlers.MakeSecretHandler(functionNamespace, kube),
		LogHandler:           logs.NewLogHandlerFunc(faasnetesk8s.NewLogRequestor(kube, functionNamespace), bootstrapConfig.WriteTimeout),
		ListNamespaceHandler: handlers.MakeNamespacesLister(functionNamespace, clusterRole, kube),
	}

	if pprof == "true" {
		bootstrap.Router().PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	bootstrap.Router().Path("/metrics").Handler(promhttp.Handler())

	glog.Infof("Using namespace '%s'", functionNamespace)

	return &Server{
		BootstrapConfig:   &bootstrapConfig,
		BootstrapHandlers: &bootstrapHandlers,
	}
}

type Server struct {
	BootstrapHandlers *types.FaaSHandlers
	BootstrapConfig   *types.FaaSConfig
}

// Start begins the server
func (s *Server) Start() {
	glog.Infof("Starting HTTP server on port %d", *s.BootstrapConfig.TCPPort)

	bootstrap.Serve(s.BootstrapHandlers, s.BootstrapConfig)
}
