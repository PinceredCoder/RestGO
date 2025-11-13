package handlers

import (
	"io"
	"net/http"

	tasks "github.com/PinceredCoder/RestGo/api/proto/v1"
	"github.com/PinceredCoder/RestGo/internal/errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskHandler struct {
	tasks []*tasks.Task
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		tasks: make([]*tasks.Task, 0),
	}
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := &tasks.ListTasksResponse{
		Tasks: h.tasks,
	}

	data, err := protojson.Marshal(response)
	if err != nil {
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to encode response"))
		return
	}

	w.Write(data)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Failed to read request body"))
		return
	}

	var req tasks.CreateTaskRequest
	if err := protojson.Unmarshal(data, &req); err != nil {
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid JSON format"))
		return
	}

	if err := req.Validate(); err != nil {
		apiErr := h.convertValidationError(err)
		errors.RespondWithError(w, http.StatusBadRequest, apiErr)
		return
	}

	now := timestamppb.Now()

	task := &tasks.Task{
		Id:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	h.tasks = append(h.tasks, task)

	response := &tasks.GetTaskResponse{
		Task: task,
	}

	data, err = protojson.Marshal(response)
	if err != nil {
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to encode response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	for _, task := range h.tasks {
		if task.Id == id {
			response := &tasks.GetTaskResponse{
				Task: task,
			}

			data, err := protojson.Marshal(response)
			if err != nil {
				errors.RespondWithError(w, http.StatusInternalServerError,
					errors.NewInternalError("Failed to encode response"))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			return
		}
	}

	errors.RespondWithError(w, http.StatusNotFound,
		errors.NewNotFoundError("Task not found"))
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	data, err := io.ReadAll(r.Body)
	if err != nil {
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Failed to read request body"))
		return
	}

	var req tasks.UpdateTaskRequest
	if err := protojson.Unmarshal(data, &req); err != nil {
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid JSON format"))
		return
	}

	if err := req.Validate(); err != nil {
		apiErr := h.convertValidationError(err)
		errors.RespondWithError(w, http.StatusBadRequest, apiErr)
		return
	}

	for i, task := range h.tasks {
		if task.Id == id {
			h.tasks[i].Title = req.Title
			h.tasks[i].Description = req.Description

			if req.Completed != nil {
				h.tasks[i].Completed = *req.Completed
			}

			h.tasks[i].UpdatedAt = timestamppb.Now()

			response := &tasks.GetTaskResponse{
				Task: h.tasks[i],
			}

			data, err := protojson.Marshal(response)
			if err != nil {
				errors.RespondWithError(w, http.StatusInternalServerError,
					errors.NewInternalError("Failed to encode response"))
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(data)
			return
		}
	}

	errors.RespondWithError(w, http.StatusNotFound,
		errors.NewNotFoundError("Task not found"))
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	for i, task := range h.tasks {
		if task.Id == id {
			h.tasks = append(h.tasks[:i], h.tasks[i+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}

	errors.RespondWithError(w, http.StatusNotFound,
		errors.NewNotFoundError("Task not found"))
}
