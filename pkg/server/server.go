package server

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"time"

	"github.com/openfaas/faas-netes/k8s"
	faasnetesk8s "github.com/openfaas/faas-netes/k8s"
	clientset "github.com/openfaas/faas-netes/pkg/client/clientset/versioned"
	bootstrap "github.com/openfaas/faas-provider"

	"github.com/openfaas/faas-provider/logs"
	"github.com/openfaas/faas-provider/proxy"
	"github.com/openfaas/faas-provider/types"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	appsinformer "k8s.io/client-go/informers/apps/v1"
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
	deploymentsInformer appsinformer.DeploymentInformer) *Server {

	functionNamespace := "openfaas-fn"
	if namespace, exists := os.LookupEnv("function_namespace"); exists {
		functionNamespace = namespace
	}

	port := defaultHTTPPort
	if portVal, exists := os.LookupEnv("port"); exists {
		parsedVal, parseErr := strconv.Atoi(portVal)
		if parseErr == nil && parsedVal > 0 {
			port = parsedVal
		}
	}

	readTimeout := defaultReadTimeout
	if val, exists := os.LookupEnv("read_timeout"); exists {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal > 0 {
			readTimeout = parsedVal
		}
	}

	writeTimeout := defaultWriteTimeout
	if val, exists := os.LookupEnv("write_timeout"); exists {
		parsedVal, parseErr := strconv.Atoi(val)
		if parseErr == nil && parsedVal > 0 {
			writeTimeout = parsedVal
		}
	}

	pprof := "false"
	if val, exists := os.LookupEnv("pprof"); exists {
		pprof = val
	}

	lister := endpointsInformer.Lister()
	functionLookup := k8s.NewFunctionLookup(functionNamespace, lister)

	deploymentLister := deploymentsInformer.Lister().Deployments(functionNamespace)
	bootstrapConfig := types.FaaSConfig{
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		TCPPort:      &port,
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
		SecretHandler:        makeSecretHandler(functionNamespace, kube),
		ListNamespaceHandler: makeListNamespaceHandler(functionNamespace),
		LogHandler:           logs.NewLogHandlerFunc(faasnetesk8s.NewLogRequestor(kube, functionNamespace), bootstrapConfig.WriteTimeout),
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
