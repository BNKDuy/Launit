package api

import (
	"encoding/json"
	"net/http"
	"orchestrator/internal/engine"
	"orchestrator/internal/store"
	"strings"
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
	mux.HandleFunc("POST /api/compute/upload", h.HandleUpload)
	mux.HandleFunc("POST /api/compute", h.HandleCreate)
	mux.HandleFunc("DELETE /api/compute/{id}", h.HandleDelete)
}

type UploadRequest struct {
	Username     string `json:"username"`
	FunctionName string `json:"name"`
}

func (h *ComputeHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	var req UploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
	}

	name := req.Username
	username := req.FunctionName
	key := h.store.GetKey(username, name)

	presignedUrl, err := h.store.GenerateUploadURL(r.Context(), key)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"upload_url": presignedUrl,
	})
}

type CreateRequest struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Memory   int32  `json:"memory"`
	Runtime  string `json:"runtime"`
	Timeout  int32  `json:"timeout"`
	UploadID string `json:"upload_id"`
}

func (h *ComputeHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		http.Error(w, "The name of the function cannot be empty", http.StatusBadRequest)
		return
	}

	// Make sure the memory is within limit
	memory := req.Memory
	if memory < engine.MIN_MEMORY || memory > engine.MAX_MEMORY {
		if memory < engine.MIN_MEMORY {
			http.Error(w, "Invalid function memory: The memory cannot be less than 128MB", http.StatusBadRequest)
		} else {
			http.Error(w, "Invalid function memory: The memory cannot exceed 10GB", http.StatusBadRequest)
		}

		return
	}

	runtime := req.Runtime
	timeout := req.Timeout
	username := req.Username

	uri := h.store.GetKey(username, name)

	url, err := h.engine.Create(r.Context(), name, memory, runtime, timeout, uri)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(url))
}

type DeleteRequest struct {
	FunctionName string `json:"function_name"`
	Username     string `json:"username"`
}

func (h *ComputeHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.FunctionName)
	if len(name) == 0 || len(name) > 16 {
		if len(name) == 0 {
			http.Error(w, "Function name cannot be empty", http.StatusBadRequest)
		} else {
			http.Error(w, "Function name cannot be longer than 16 characters", http.StatusBadRequest)
		}
		return
	}

	if err := h.engine.Delete(r.Context(), name); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to delete function!"))
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Function successfully taken down!"))
}
