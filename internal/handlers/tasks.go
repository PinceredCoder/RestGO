package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/PinceredCoder/RestGo/internal/errors"
	"github.com/PinceredCoder/RestGo/internal/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TaskHandler struct {
	tasks    []models.Task
	validate *validator.Validate
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		tasks:    make([]models.Task, 0),
		validate: validator.New(),
	}
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if h.tasks == nil {
		h.tasks = []models.Task{}
	}

	if err := json.NewEncoder(w).Encode(h.tasks); err != nil {
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to encode response"))
		return
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid JSON format"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := h.extractValidationErrors(err)
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewValidationError("Validation failed", validationErrors))
		return
	}

	var now = time.Now()

	task := models.Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	h.tasks = append(h.tasks, task)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	for _, task := range h.tasks {
		if task.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(task)
			return
		}
	}

	errors.RespondWithError(w, http.StatusNotFound,
		errors.NewNotFoundError("Task not found"))
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req models.UpdateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid JSON format"))
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := h.extractValidationErrors(err)
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewValidationError("Validation failed", validationErrors))
		return
	}
	for i, task := range h.tasks {
		if task.ID == id {
			h.tasks[i].Title = req.Title
			h.tasks[i].Description = req.Description

			if req.Completed != nil {
				h.tasks[i].Completed = *req.Completed
			}

			h.tasks[i].UpdatedAt = time.Now()

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(h.tasks[i])
			return
		}
	}

	errors.RespondWithError(w, http.StatusNotFound,
		errors.NewNotFoundError("Task not found"))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	for i, task := range h.tasks {
		if task.ID == id {
			h.tasks = append(h.tasks[:i], h.tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	errors.RespondWithError(w, http.StatusNotFound,
		errors.NewNotFoundError("Task not found"))
}

func (h *TaskHandler) extractValidationErrors(err error) []errors.ValidationErrorDetail {
	var details []errors.ValidationErrorDetail

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			detail := errors.ValidationErrorDetail{
				Field:   fieldError.Field(),
				Message: h.getValidationMessage(fieldError),
			}
			details = append(details, detail)
		}
	}

	return details
}

func (h *TaskHandler) getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Value is too short (minimum " + fe.Param() + " characters)"
	case "max":
		return "Value is too long (maximum " + fe.Param() + " characters)"
	default:
		return "Invalid value"
	}
}
