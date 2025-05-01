package http

import (
	"net/http"

	"github.com/singl3focus/load_balancer-go/internal/adapters/primary/http/handlers"
)

func NewHandler(
	lb *handlers.LoadBalancer,
	ch *handlers.ClientHandler,
) http.Handler {
	mux := http.NewServeMux()

	// api 
	mux.HandleFunc("POST /clients", ch.AddConfig)
	mux.HandleFunc("GET /clients", ch.GetConfig)
	mux.HandleFunc("PATCH /clients", ch.UpdateConfig)
	mux.HandleFunc("DELETE /clients", ch.DeleteConfig)

	// balancer
	mux.Handle("/", lb)

	return mux
}
