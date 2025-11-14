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
	tasks map[string]*tasks.Task
}

func NewTaskHandler() *TaskHandler {
	return &TaskHandler{
		tasks: make(map[string]*tasks.Task),
	}
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	taskList := make([]*tasks.Task, 0, len(h.tasks))
	for _, task := range h.tasks {
		taskList = append(taskList, task)
	}

	response := &tasks.ListTasksResponse{
		Tasks: taskList,
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

	h.tasks[task.Id] = task

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

	task, exists := h.tasks[id]
	if !exists {
		errors.RespondWithError(w, http.StatusNotFound,
			errors.NewNotFoundError("Task not found"))
		return
	}

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

	task, exists := h.tasks[id]
	if !exists {
		errors.RespondWithError(w, http.StatusNotFound,
			errors.NewNotFoundError("Task not found"))
		return
	}

	task.Title = req.Title
	task.Description = req.Description

	if req.Completed != nil {
		task.Completed = *req.Completed
	}

	task.UpdatedAt = timestamppb.Now()

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
	w.Write(data)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if _, exists := h.tasks[id]; !exists {
		errors.RespondWithError(w, http.StatusNotFound,
			errors.NewNotFoundError("Task not found"))
		return
	}

	delete(h.tasks, id)
	w.WriteHeader(http.StatusNoContent)
}
