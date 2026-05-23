package handlers

import (
	"encoding/json"
	"errors"
	"go-rest/internal/database"
	"go-rest/internal/models"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type Handler struct {
	store *database.TaskStore
}

func NewHandler(store *database.TaskStore) *Handler {
	return &Handler{store: store}
}

// Utils
func respondWithJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		// Заголовки и статус уже отправлены — ответ не исправить, только логируем.
		log.Printf("encode response: %v", err)
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{"error": message})
}

// taskIDFromRequest достаёт {id} из пути, который ServeMux уже распарсил.
func taskIDFromRequest(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

// /Utils

func (h *Handler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.store.GetAll()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Ошибка получения задач")
		return
	}

	respondWithJSON(w, http.StatusOK, tasks)
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	id, err := taskIDFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный id задачи")
		return
	}

	task, err := h.store.GetByID(id)
	if err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	var input models.CreateTaskInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректно отправлены данные")
		return
	}

	if strings.TrimSpace(input.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Title required")
		return
	}

	task, err := h.store.Create(input)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, task)
}

func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	id, err := taskIDFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный id задачи")
		return
	}

	var input models.UpdateTaskInput

	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректно отправлены данные")
		return
	}

	if input.Title != nil && strings.TrimSpace(*input.Title) == "" {
		respondWithError(w, http.StatusBadRequest, "Title required")
		return
	}

	task, err := h.store.Update(id, input)
	if err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}

		return
	}

	respondWithJSON(w, http.StatusOK, task)
}

func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	id, err := taskIDFromRequest(r)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Некорректный id задачи")
		return
	}

	err = h.store.Delete(id)
	if err != nil {
		if errors.Is(err, database.ErrTaskNotFound) {
			respondWithError(w, http.StatusNotFound, err.Error())
		} else {
			respondWithError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	respondWithJSON(w, http.StatusNoContent, map[string]string{"result": "success"})
}
