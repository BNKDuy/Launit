package api

import (
	"net/http"
	"orchestrator/internal/engine"
	"orchestrator/internal/store"
)

type ComputeHandler struct {
	engine engine.Engine
	store  store.Store
}

func NewComputeHandler(engine engine.Engine, store store.Store) *ComputeHandler {
	return &ComputeHandler{
		engine: engine,
		store:  store,
	}
}

func (h *ComputeHandler) RegisterRoutes(mux *http.ServeMux) {
	// mux.HandleFunc("POST /api/compute/upload", )
	// mux.HandleFunc("POST /api/compute", )
	// mux.HandleFunc("DELETE /api/compute/{id}", )
}
