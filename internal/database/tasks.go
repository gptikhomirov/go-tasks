package database

import (
	"database/sql"
	"errors"
	"fmt"
	"go-rest/internal/models"
	"time"

	"github.com/jmoiron/sqlx"
)

var ErrTaskNotFound = errors.New("task not found")

type TaskStore struct {
	db *sqlx.DB
}

func NewTaskStore(db *sqlx.DB) *TaskStore {
	return &TaskStore{db: db}
}

func (s *TaskStore) GetAll(params models.GetAllParams) ([]models.TaskListItem, error) {
	tasks := []models.TaskListItem{}

	var (
		args  []any
		where string
	)

	if params.Search != "" {
		where = "WHERE title ILIKE $1 OR description ILIKE $1"
		args = append(args, "%"+params.Search+"%")
	}

	query := `
    	SELECT title, description, completed, created_at
    	FROM tasks
	` + where + `
    	ORDER BY created_at DESC
	`

	err := s.db.Select(&tasks, query, params.Search)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s *TaskStore) GetByID(id int) (*models.Task, error) {
	var task models.Task

	query := `
		SELECT id, title, description, completed, created_at, updated_at
		FROM tasks
		WHERE id = $1;
	`

	err := s.db.Get(&task, query, id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("task with id %d: %w", id, ErrTaskNotFound)
	}

	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *TaskStore) Create(input models.CreateTaskInput) (*models.Task, error) {
	var task models.Task

	query := `
		INSERT INTO tasks (title, description, completed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, title, description, completed, created_at, updated_at;
	`
	now := time.Now()

	err := s.db.QueryRowx(query, input.Title, input.Description, input.Completed, now, now).StructScan(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (s *TaskStore) Update(id int, input models.UpdateTaskInput) (*models.Task, error) {
	task, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Completed != nil {
		task.Completed = *input.Completed
	}
	task.UpdatedAt = time.Now()

	query := `
		UPDATE tasks
		SET title = $1, description = $2, completed = $3, updated_at = $4
		WHERE id = $5
		RETURNING id, title, description, completed, created_at, updated_at;
	`

	var updatedTask models.Task
	err = s.db.QueryRowx(query, task.Title, task.Description, task.Completed, task.UpdatedAt, task.ID).StructScan(&updatedTask)
	if err != nil {
		return nil, err
	}

	return &updatedTask, nil
}

func (s *TaskStore) Delete(id int) error {
	query := `DELETE FROM tasks WHERE id = $1;`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return fmt.Errorf("task with id %d: %w", id, ErrTaskNotFound)
	}

	return nil
}
