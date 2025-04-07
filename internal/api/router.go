package api

import (
	"net/http"

	"iot-bridge/internal/api/handlers"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Scan-related APIs
	r.Route("/scan", func(r chi.Router) {
		r.Post("/", handlers.StartScan)
		r.Get("/results", handlers.GetScanResults)
	})

	// Device APIs
	r.Route("/devices", func(r chi.Router) {
		r.Get("/", handlers.GetDevices)
		r.Post("/", handlers.AddDevice)
		r.Post("/from-scan", handlers.AddDeviceFromScan)

		r.Get("/{id}", handlers.GetDeviceByID)
		r.Patch("/{id}", handlers.PatchDevice)
		r.Delete("/{id}", handlers.DeleteDevice)

		r.Get("/{id}/capabilities", handlers.GetCapabilities)
		r.Post("/{id}/capabilities/{capability}", handlers.InvokeCapability)
		r.Post("/{id}/capabilities", handlers.UpdateCapabilities)
	})

	// LLM interaction (POST) and optional browser support
	r.Post("/llm", handlers.HandleLLMRequest)
	r.Get("/llm", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	// Web UI - Serve index.html from /web
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	// Static assets (optional: CSS, JS, etc.)
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir("./web"))))

	return r
}
