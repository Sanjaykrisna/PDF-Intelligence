package main

import (
	"fmt"
	"log"
	"net/http"

	"backend/config"
	"backend/handlers"
	"backend/repositories"
	"backend/services"
	"backend/utils"
)

func main() {
	// Initialize unique log file in "log" folder for all detailed runtime traces
	lLogFile, lLogPath, lErr := utils.InitFileLogger()
	if lErr != nil {
		fmt.Printf("[LOG ERROR] Failed to initialize file logger: %v\n", lErr)
	} else {
		defer lLogFile.Close()
		fmt.Printf("[LOG] Runtime logs recording to: %s\n", lLogPath)
	}

	lDebug := new(utils.HelperStruct)
	lDebug.Init()
	lDebug.Log(utils.Statement, "main(+)")

	// 1. Load configuration
	lErr = config.LoadConfig("toml/config.toml")
	if lErr != nil {
		fmt.Printf("[CONFIG ERROR] Failed to load config: %v\n", lErr)
		log.Fatalf("Failed to load config: %v", lErr)
		return
	}
	fmt.Printf("[CONFIG] Configuration loaded from toml/config.toml (Port: %d, DB: %s, AI: %s)\n",
		config.AppConfig.Server.Port, config.AppConfig.Database.DBName, config.AppConfig.AI.Model)

	// 2. Open DB Connection & defer close (TradeApp standard pattern)
	lErr = repositories.InitDB(lDebug)
	if lErr != nil {
		fmt.Printf("[DATABASE ERROR] Failed to initialize database: %v\n", lErr)
		log.Fatalf("Failed to initialize database: %v", lErr)
		return
	}
	defer repositories.CloseDB()
	fmt.Printf("[DATABASE] PostgreSQL connected successfully (%s:%d/%s)\n",
		config.AppConfig.Database.Host, config.AppConfig.Database.Port, config.AppConfig.Database.DBName)

	// 3. Register HTTP API Endpoints directly (Upload, Ask, Summary, Insights, Delete, List)
	http.HandleFunc("/api/system/info", handlers.GetSystemInfo)

	// Upload
	http.HandleFunc("/api/documents/upload", handlers.UploadDocument)

	// Ask / Question
	http.HandleFunc("/api/documents/ask", handlers.Ask)
	http.HandleFunc("/api/documents/{id}/ask", handlers.Ask)

	// Summary
	http.HandleFunc("/api/documents/summary", handlers.GetSummary)
	http.HandleFunc("/api/documents/{id}/summary", handlers.GetSummary)

	// Insights & Audits
	http.HandleFunc("/api/documents/insights", handlers.GetInsights)
	http.HandleFunc("/api/documents/{id}/insights", handlers.GetInsights)

	// Latest Document
	http.HandleFunc("/api/documents/latest", handlers.GetLatestDocument)

	// Delete
	http.HandleFunc("/api/documents/delete", handlers.DeleteDocument)
	http.HandleFunc("/api/documents/{id}/delete", handlers.DeleteDocument)
	http.HandleFunc("/api/documents/{id}", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete || r.Method == http.MethodPost || r.Method == http.MethodOptions {
			handlers.DeleteDocument(w, r)
			return
		}
		http.NotFound(w, r)
	})

	// List & Document collection
	http.HandleFunc("/api/documents", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete || r.Method == http.MethodPost {
			handlers.DeleteDocument(w, r)
			return
		}
		handlers.GetDocuments(w, r)
	})

	// 4. Background Schedulers (4 times daily)
	services.StartDailyFlushScheduler(lDebug)

	// 5. Start Server
	lPort := config.AppConfig.Server.Port
	if lPort == 0 {
		lPort = 8080
	}

	fmt.Printf("server running in port:%d\n", lPort)
	lAddr := fmt.Sprintf(":%d", lPort)
	lDebug.Log(utils.Statement, "Server starting on "+lAddr)

	lErr = http.ListenAndServe(lAddr, nil)
	if lErr != nil {
		log.Println("Server failed:", lErr)
	}

	lDebug.Log(utils.Statement, "main(-)")
}
