package main

import (
	"go-rest/internal/database"
	"go-rest/internal/handlers"
	"log"
	"net/http"
	"os"
)

const (
	defaultPort = "8080"
)

func getPort() string {
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		return defaultPort
	}

	return serverPort
}

// TODO: env
func main() {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		databaseUrl = "postgres://tasks_user:password@localhost:5432/tasks_db?sslmode=disable"
	}

	db, err := database.Connect(databaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Printf("Connected to database %s", databaseUrl)

	taskStore := database.NewTaskStore(db)
	handler := handlers.NewHandler(taskStore)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /tasks", handler.GetAllTasks)
	mux.HandleFunc("POST /tasks", handler.CreateTask)
	mux.HandleFunc("GET /tasks/{id}", handler.GetTask)
	mux.HandleFunc("PUT /tasks/{id}", handler.UpdateTask)
	mux.HandleFunc("DELETE /tasks/{id}", handler.DeleteTask)

	loggedMux := loggingMiddleware(mux)

	// TODO: cors middleware

	serverPort := getPort()
	log.Printf("Starting server on port %s", serverPort)
	serverAddr := ":" + serverPort

	err = http.ListenAndServe(serverAddr, loggedMux)
	if err != nil {
		log.Fatal(err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
