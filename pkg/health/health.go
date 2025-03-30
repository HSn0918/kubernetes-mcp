package health

import (
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/hsn0918/kubernetes-mcp/pkg/logger"
)

var (
	isReady int32 // Atomic boolean: 0 = not ready, 1 = ready
	log     logger.Logger
)

// SetReady marks the application as ready.
func SetReady() {
	atomic.StoreInt32(&isReady, 1)
	if log != nil {
		log.Info("Application marked as ready for health checks")
	}
}

// SetNotReady marks the application as not ready.
func SetNotReady() {
	atomic.StoreInt32(&isReady, 0)
	if log != nil {
		log.Warn("Application marked as not ready for health checks")
	}
}

// healthzHandler handles liveness probes.
// Checks if the process is running. A simple 200 OK is usually sufficient.
func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "OK")
}

// readyzHandler handles readiness probes.
// Checks if the application is ready to serve requests.
// Here we check our atomic 'isReady' flag.
// A more complex check could verify dependencies like the K8s client.
func readyzHandler(w http.ResponseWriter, r *http.Request) {
	if atomic.LoadInt32(&isReady) == 1 {
		// Optional: Add checks for critical dependencies like Kubernetes client connection
		// k8sClient := client.GetClient() // Get the initialized client
		// if k8sClient == nil {
		// 	http.Error(w, "Kubernetes client not initialized", http.StatusServiceUnavailable)
		//  log.Warn("Readiness check failed: K8s client not initialized")
		// 	return
		// }
		// Add a simple check, e.g., try listing namespaces with a timeout (be careful not to overload API server)

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "OK")
	} else {
		http.Error(w, "Service not ready", http.StatusServiceUnavailable)
		if log != nil {
			log.Warn("Readiness check failed: Service not marked as ready")
		}
	}
}

// StartHealthServer starts a simple HTTP server for health checks on a separate port.
func StartHealthServer(port int, logger logger.Logger) {
	log = logger // Store logger for handlers
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", healthzHandler)
	mux.HandleFunc("/readyz", readyzHandler)

	healthServer := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	log.Info("Starting health check server", "port", port)
	// Run the server in a separate goroutine so it doesn't block the main application
	go func() {
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Health check server failed", "error", err)
		}
	}()

	// Initially mark as not ready until main server components are up
	SetNotReady()
}
