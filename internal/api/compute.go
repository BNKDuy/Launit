package api

import (
	"context"
	"encoding/json"
	"net/http"
	"orchestrator/internal/authenticator"
	"orchestrator/internal/engine"
	"orchestrator/internal/store"
	"strings"
)

type contextKey string

const (
	UsernameKey           contextKey = "username"
	ComputeUploadEndpoint string     = "/api/compute/upload"
	ComputeCreateEndpoint string     = "/api/compute"
	ComputeDeleteEndpoint string     = "/api/compute/{id}"
)

type ComputeHandler struct {
	engine        engine.Engine
	store         store.Store
	authenticator authenticator.Authenticator
}

func NewComputeHandler(e engine.Engine, s store.Store, a authenticator.Authenticator) *ComputeHandler {
	return &ComputeHandler{
		engine:        e,
		store:         s,
		authenticator: a,
	}
}

func (h *ComputeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.Handle("POST "+ComputeUploadEndpoint, h.AuthMiddleware(http.HandlerFunc(h.HandleUpload)))
	mux.Handle("POST "+ComputeCreateEndpoint, h.AuthMiddleware(http.HandlerFunc(h.HandleCreate)))
	mux.Handle("DELETE "+ComputeDeleteEndpoint, h.AuthMiddleware(http.HandlerFunc(h.HandleDelete)))
}

func (h *ComputeHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: Missing Authorization header", http.StatusUnauthorized)
			return
		}

		// Parse out the token from the "Bearer <token>" format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "Unauthorized: Invalid Authorization format. Use 'Bearer <token>'", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		username, err := h.authenticator.Authenticate(r.Context(), token)
		if err != nil {
			http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Inject the username into the request Context
		ctx := context.WithValue(r.Context(), UsernameKey, username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

type uploadRequest struct {
	FunctionName string `json:"1"`
}

func newUploadRequest(functionName string) *uploadRequest {
	return &uploadRequest{
		FunctionName: functionName,
	}
}

func (h *ComputeHandler) HandleUpload(w http.ResponseWriter, r *http.Request) {
	var req uploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	name := req.FunctionName
	username := h.getUsernameFromContext(r.Context())
	key := h.store.GetKey(username, name)

	presignedUrl, err := h.store.GenerateUploadURL(r.Context(), key)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := newPresignedUrlResponse(presignedUrl)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

type CreateRequest struct {
	FunctionName string `json:"1"`
	Size         string `json:"2"`
	Runtime      string `json:"3"`
	Timeout      int32  `json:"4"`
}

func newCreateRequest(functionName, size, runtime string, timeout int32) *CreateRequest {
	return &CreateRequest{
		FunctionName: functionName,
		Size:         size,
		Runtime:      runtime,
		Timeout:      timeout,
	}
}

func (h *ComputeHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	name := strings.TrimSpace(req.FunctionName)
	if name == "" {
		http.Error(w, "The name of the function cannot be empty", http.StatusBadRequest)
		return
	}

	size := req.Size
	runtime := req.Runtime
	timeout := req.Timeout

	username := h.getUsernameFromContext(r.Context())

	uri := h.store.GetKey(username, name)

	url, err := h.engine.Create(r.Context(), name, size, runtime, timeout, uri)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	res := newCreateResponse(url)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *ComputeHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimSpace(r.PathValue("id"))
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
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Function successfully taken down!"))
}

func (h *ComputeHandler) getUsernameFromContext(ctx context.Context) string {
	if username, ok := ctx.Value(UsernameKey).(string); ok {
		return username
	}
	return ""
}
