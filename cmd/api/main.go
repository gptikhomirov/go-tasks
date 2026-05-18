package main

import (
	"go-start/internal/database"
	"go-start/internal/handlers"
	"log"
	"net/http"
	"os"
)

func main() {
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		databaseUrl = "postgres://tasks_user:password@localhost:5432/tasks_db?sslmode=disable"
	}

	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}
	log.Printf("Starting server on port %s", serverPort)

	db, err := database.Connect(databaseUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Printf("Connected to database %s", databaseUrl)

	taskStore := database.NewTaskStore(db)
	handler := handlers.NewHandler(taskStore)

	mux := http.NewServeMux()
	mux.HandleFunc("/tasks", methodHandler(handler.GetAllTasks, http.MethodGet))
	mux.HandleFunc("/tasks/create", methodHandler(handler.CreateTask, http.MethodPost))
	//mux.HandleFunc("/tasks/:id", methodHandler(handler.GetTask, http.MethodGet))
	//mux.HandleFunc("/tasks/:id", methodHandler(handler.UpdateTask, http.MethodPut))
	//mux.HandleFunc("/tasks/:id", methodHandler(handler.DeleteTask, http.MethodDelete))

	mux.HandleFunc("/tasks/", taskIDHandler(handler))

	loggedMux := loggingMiddleware(mux)

	// TODO: cors middleware

	serverAddr := ":" + serverPort

	err = http.ListenAndServe(serverAddr, loggedMux)
	if err != nil {
		log.Fatal(err)
	}

}

func methodHandler(handlerFunc http.HandlerFunc, allowedMethod string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != allowedMethod {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}

		handlerFunc(w, r)
	}
}

func taskIDHandler(handler *handlers.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.GetTask(w, r)
		case http.MethodPut:
			handler.UpdateTask(w, r)
		case http.MethodDelete:
			handler.DeleteTask(w, r)
		default:
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		}
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
