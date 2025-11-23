package handlers

import (
	"io"
	"log/slog"
	"net/http"

	tasks "github.com/PinceredCoder/restGo/api/proto/v1"
	"github.com/PinceredCoder/restGo/internal/database"
	"github.com/PinceredCoder/restGo/internal/errors"
	"github.com/PinceredCoder/restGo/internal/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type TaskHandler struct {
	db     database.Database
	logger *slog.Logger
}

func NewTaskHandler(db database.Database, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{
		db:     db,
		logger: logger,
	}
}

func (h *TaskHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	h.logger.Info("Fetching all tasks")

	taskList, err := h.db.GetTaskRepository().FindAll(r.Context())

	if err != nil {
		h.logger.Error("Failed to retrieve tasks from database", "error", err)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to retrieve tasks"))
		return
	}

	h.logger.Info("Successfully retrieved tasks", "count", len(taskList))

	response := &tasks.ListTasksResponse{
		Tasks: helpers.Map(taskList, func(t *database.Task) *tasks.Task { return t.ToProto() }),
	}

	data, err := protojson.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal response", "error", err)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to encode response"))
		return
	}

	w.Write(data)
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Creating new task")

	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Warn("Failed to read request body", "error", err)
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Failed to read request body"))
		return
	}

	var req tasks.CreateTaskRequest
	if err := protojson.Unmarshal(data, &req); err != nil {
		h.logger.Warn("Invalid JSON format in request", "error", err)
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid JSON format"))
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Warn("Validation failed for create request", "error", err)
		apiErr := h.convertValidationError(err)
		errors.RespondWithError(w, http.StatusBadRequest, apiErr)
		return
	}

	now := timestamppb.Now().AsTime().Unix()
	taskID := uuid.New()

	taskDb := &database.Task{
		ID:          taskID,
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.db.GetTaskRepository().Create(r.Context(), taskDb); err != nil {
		h.logger.Error("Failed to create task in database", "error", err, "task_id", taskID)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to create task"))
		return
	}

	h.logger.Info("Task created successfully", "task_id", taskID, "title", taskDb.Title)

	response := &tasks.GetTaskResponse{
		Task: taskDb.ToProto(),
	}

	data, err = protojson.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal create response", "error", err, "task_id", taskID)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to encode response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(data)
}

func (h *TaskHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("Invalid task ID format", "id", idStr)
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid task ID format"))
		return
	}

	h.logger.Info("Fetching task by ID", "task_id", id)

	taskDb, err := h.db.GetTaskRepository().FindByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to retrieve task from database", "error", err, "task_id", id)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to retrieve task"))
		return
	}

	if taskDb == nil {
		h.logger.Info("Task not found", "task_id", id)
		errors.RespondWithError(w, http.StatusNotFound,
			errors.NewNotFoundError("Task not found"))
		return
	}

	h.logger.Info("Task retrieved successfully", "task_id", id)

	response := &tasks.GetTaskResponse{
		Task: taskDb.ToProto(),
	}

	data, err := protojson.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal GetByID response", "error", err, "task_id", id)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to encode response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("Invalid task ID format for update", "id", idStr)
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid task ID format"))
		return
	}

	h.logger.Info("Updating task", "task_id", id)

	data, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Warn("Failed to read update request body", "error", err, "task_id", id)
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Failed to read request body"))
		return
	}

	var req tasks.UpdateTaskRequest
	if err := protojson.Unmarshal(data, &req); err != nil {
		h.logger.Warn("Invalid JSON format in update request", "error", err, "task_id", id)
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid JSON format"))
		return
	}

	if err := req.Validate(); err != nil {
		h.logger.Warn("Validation failed for update request", "error", err, "task_id", id)
		apiErr := h.convertValidationError(err)
		errors.RespondWithError(w, http.StatusBadRequest, apiErr)
		return
	}

	task, err := h.db.GetTaskRepository().FindByID(r.Context(), id)
	if err != nil {
		h.logger.Error("Failed to retrieve task for update", "error", err, "task_id", id)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to retrieve task"))
		return
	}
	if task == nil {
		h.logger.Info("Task not found for update", "task_id", id)
		errors.RespondWithError(w, http.StatusNotFound,
			errors.NewNotFoundError("Task not found"))
		return
	}

	task.Title = req.Title
	task.Description = req.Description

	if req.Completed != nil {
		task.Completed = *req.Completed
	}

	task.UpdatedAt = timestamppb.Now().AsTime().Unix()

	if err := h.db.GetTaskRepository().Update(r.Context(), id, task); err != nil {
		h.logger.Error("Failed to update task in database", "error", err, "task_id", id)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to update task"))
		return
	}

	h.logger.Info("Task updated successfully", "task_id", id, "title", task.Title)

	response := &tasks.GetTaskResponse{
		Task: task.ToProto(),
	}

	data, err = protojson.Marshal(response)
	if err != nil {
		h.logger.Error("Failed to marshal update response", "error", err, "task_id", id)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to encode response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	id, err := uuid.Parse(idStr)
	if err != nil {
		h.logger.Warn("Invalid task ID format for delete", "id", idStr)
		errors.RespondWithError(w, http.StatusBadRequest,
			errors.NewBadRequestError("Invalid task ID format"))
		return
	}

	h.logger.Info("Deleting task", "task_id", id)

	if err := h.db.GetTaskRepository().Delete(r.Context(), id); err != nil {
		h.logger.Error("Failed to delete task from database", "error", err, "task_id", id)
		errors.RespondWithError(w, http.StatusInternalServerError,
			errors.NewInternalError("Failed to delete task"))
		return
	}

	h.logger.Info("Task deleted successfully", "task_id", id)

	w.WriteHeader(http.StatusNoContent)
}
