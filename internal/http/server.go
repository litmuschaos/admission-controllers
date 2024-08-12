package http

import (
	"crypto/tls"
	"fmt"
	"github.com/litmuschaos/admission-controller/pkg/clients"
	"net/http"

	"github.com/litmuschaos/admission-controller/internal/pods"
)

// NewServer creates and return a http.Server
func NewServer(port string, clients clients.ClientSets, sCert *tls.Certificate) *http.Server {
	// Instances hooks
	podsValidation := pods.NewValidationHook(clients)

	// Routers
	ah := newAdmissionHandler()
	mux := http.NewServeMux()
	mux.Handle("/healthz", healthz())
	mux.Handle("/validate/pods", ah.Serve(podsValidation))

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	if sCert != nil {
		httpServer.TLSConfig = &tls.Config{Certificates: []tls.Certificate{*sCert}}
	}

	return httpServer
}
