package handlers

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	tasks "github.com/PinceredCoder/restGo/api/proto/v1"
	"github.com/PinceredCoder/restGo/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
)

// setupRouter creates a chi router with task handler routes
// This allows us to test with URL parameters properly
func setupRouter() (*chi.Mux, *TaskHandler) {
	r := chi.NewRouter()
	mockDB := NewMockDatabase()
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Only log errors in tests
	}))
	h := NewTaskHandler(mockDB, logger)

	r.Get("/api/v1/tasks", h.GetAll)
	r.Post("/api/v1/tasks", h.Create)
	r.Get("/api/v1/tasks/{id}", h.GetByID)
	r.Put("/api/v1/tasks/{id}", h.Update)
	r.Delete("/api/v1/tasks/{id}", h.Delete)

	return r, h
}

// TestIntegrationCreate tests full create flow
func TestIntegrationCreate(t *testing.T) {
	router, _ := setupRouter()

	reqBody := &tasks.CreateTaskRequest{
		Title:       "Integration Test Task",
		Description: "Testing with chi router",
	}

	bodyBytes, _ := protojson.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response tasks.GetTaskResponse
	if err := protojson.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Task.Title != "Integration Test Task" {
		t.Errorf("expected title 'Integration Test Task', got '%s'", response.Task.Title)
	}
}

// TestIntegrationGetByID tests retrieving by ID with chi routing
func TestIntegrationGetByID(t *testing.T) {
	router, h := setupRouter()

	// Pre-populate a task
	taskUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440001")
	taskID := taskUUID.String()

	dbTask := &database.Task{
		ID:          taskUUID,
		Title:       "Test Task",
		Description: "Test Description",
		Completed:   false,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567890,
	}
	h.db.GetTaskRepository().Create(context.Background(), dbTask)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response tasks.GetTaskResponse
	if err := protojson.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Task.Id != taskID {
		t.Errorf("expected ID '%s', got '%s'", taskID, response.Task.Id)
	}
}

// TestIntegrationGetByIDNotFound tests 404 handling
func TestIntegrationGetByIDNotFound(t *testing.T) {
	router, _ := setupRouter()

	// Use a valid UUID that doesn't exist
	nonExistentID := "550e8400-e29b-41d4-a716-999999999999"
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+nonExistentID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

// TestIntegrationUpdate tests updating a task
func TestIntegrationUpdate(t *testing.T) {
	router, h := setupRouter()

	// Pre-populate a task
	taskUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440002")
	taskID := taskUUID.String()

	dbTask := &database.Task{
		ID:          taskUUID,
		Title:       "Original Title",
		Description: "Original Description",
		Completed:   false,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567890,
	}
	h.db.GetTaskRepository().Create(context.Background(), dbTask)

	// Update the task
	updateReq := &tasks.UpdateTaskRequest{
		Title:       "Updated Title",
		Description: "Updated Description",
	}

	bodyBytes, _ := protojson.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+taskID, bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response tasks.GetTaskResponse
	if err := protojson.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if response.Task.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", response.Task.Title)
	}

	if response.Task.Description != "Updated Description" {
		t.Errorf("expected description 'Updated Description', got '%s'", response.Task.Description)
	}
}

// TestIntegrationUpdateCompleted tests updating completion status
func TestIntegrationUpdateCompleted(t *testing.T) {
	router, h := setupRouter()

	taskUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440003")
	taskID := taskUUID.String()

	dbTask := &database.Task{
		ID:          taskUUID,
		Title:       "Task to Complete",
		Description: "Description",
		Completed:   false,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567890,
	}
	h.db.GetTaskRepository().Create(context.Background(), dbTask)

	// Mark as completed
	completed := true
	updateReq := &tasks.UpdateTaskRequest{
		Title:       "Task to Complete",
		Description: "Description",
		Completed:   &completed,
	}

	bodyBytes, _ := protojson.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+taskID, bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response tasks.GetTaskResponse
	if err := protojson.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !response.Task.Completed {
		t.Error("expected task to be completed")
	}
}

// TestIntegrationDelete tests deleting a task
func TestIntegrationDelete(t *testing.T) {
	router, h := setupRouter()

	taskUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440004")
	taskID := taskUUID.String()

	dbTask := &database.Task{
		ID:          taskUUID,
		Title:       "Task to Delete",
		Description: "Will be deleted",
		Completed:   false,
		CreatedAt:   1234567890,
		UpdatedAt:   1234567890,
	}
	h.db.GetTaskRepository().Create(context.Background(), dbTask)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+taskID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	// Verify task was deleted
	deletedTask, _ := h.db.GetTaskRepository().FindByID(context.Background(), taskUUID)
	if deletedTask != nil {
		t.Error("task should have been deleted")
	}
}

// TestIntegrationDeleteNotFound tests deleting non-existent task
func TestIntegrationDeleteNotFound(t *testing.T) {
	router, _ := setupRouter()

	// Use a valid UUID that doesn't exist
	nonExistentID := "550e8400-e29b-41d4-a716-999999999998"
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+nonExistentID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Note: Current implementation returns 204 even if not found
	// This is a known limitation that could be improved
	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}
}

// TestIntegrationFullWorkflow tests a complete CRUD workflow
func TestIntegrationFullWorkflow(t *testing.T) {
	router, _ := setupRouter()

	// 1. Create a task
	createReq := &tasks.CreateTaskRequest{
		Title:       "Workflow Task",
		Description: "Testing full workflow",
	}

	bodyBytes, _ := protojson.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("create failed with status %d", w.Code)
	}

	var createResp tasks.GetTaskResponse
	protojson.Unmarshal(w.Body.Bytes(), &createResp)
	taskID := createResp.Task.Id

	// 2. Get the task
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("get failed with status %d", w.Code)
	}

	// 3. Update the task
	completed := true
	updateReq := &tasks.UpdateTaskRequest{
		Title:       "Updated Workflow Task",
		Description: "Updated description",
		Completed:   &completed,
	}

	bodyBytes, _ = protojson.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+taskID, bytes.NewReader(bodyBytes))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("update failed with status %d", w.Code)
	}

	var updateResp tasks.GetTaskResponse
	protojson.Unmarshal(w.Body.Bytes(), &updateResp)

	if updateResp.Task.Title != "Updated Workflow Task" {
		t.Errorf("title not updated: got '%s'", updateResp.Task.Title)
	}

	if !updateResp.Task.Completed {
		t.Error("task should be marked as completed")
	}

	// 4. List all tasks (should contain our task)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list failed with status %d", w.Code)
	}

	var listResp tasks.ListTasksResponse
	protojson.Unmarshal(w.Body.Bytes(), &listResp)

	found := false
	for _, task := range listResp.Tasks {
		if task.Id == taskID {
			found = true
			break
		}
	}

	if !found {
		t.Error("created task not found in list")
	}

	// 5. Delete the task
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+taskID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("delete failed with status %d", w.Code)
	}

	// 6. Verify deletion
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Error("deleted task should return 404")
	}
}
