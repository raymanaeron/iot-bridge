package api

import (
	"net/http"

	"iot-bridge/internal/api/handlers"

	"github.com/go-chi/chi/v5"
)

func NewRouter() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	r.Route("/scan", func(r chi.Router) {
		r.Post("/", handlers.StartScan)
		r.Get("/results", handlers.GetScanResults)
	})

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

	r.Post("/llm", handlers.HandleLLMRequest)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./web/index.html")
	})

	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./web"))))

	return r
}
