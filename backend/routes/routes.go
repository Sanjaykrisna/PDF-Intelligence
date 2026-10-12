package routes

import (
	"net/http"

	"backend/handlers"
)

// SetupRoutes registers all HTTP API route handlers.
func SetupRoutes() {
	http.HandleFunc("/api/system/info", handlers.GetSystemInfo)
	http.HandleFunc("/api/documents/upload", handlers.UploadDocument)
	http.HandleFunc("/api/documents/latest", handlers.GetLatestDocument)
	http.HandleFunc("/api/documents/summary", handlers.GetSummary)
	http.HandleFunc("/api/documents/insights", handlers.GetInsights)
	http.HandleFunc("/api/documents/ask", handlers.Ask)
	http.HandleFunc("/api/documents/delete", handlers.DeleteDocument)
	http.HandleFunc("/api/documents", handlers.GetDocuments)
	http.HandleFunc("/api/documents/", handlers.DocumentsRouter)
}

