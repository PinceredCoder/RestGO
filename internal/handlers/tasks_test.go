package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	tasks "github.com/PinceredCoder/RestGo/api/proto/v1"
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Test helper: creates a task handler with some pre-populated tasks
func setupHandler() *TaskHandler {
	h := NewTaskHandler()

	// Add a test task
	now := timestamppb.Now()
	testID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	h.tasks[testID] = &tasks.Task{
		Id:          testID.String(),
		Title:       "Test Task",
		Description: "Test Description",
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	return h
}

// TestNewTaskHandler tests the constructor
func TestNewTaskHandler(t *testing.T) {
	h := NewTaskHandler()

	if h == nil {
		t.Fatal("NewTaskHandler() returned nil")
	}

	if h.tasks == nil {
		t.Error("tasks map is nil")
	}

	if len(h.tasks) != 0 {
		t.Errorf("expected empty tasks, got %d tasks", len(h.tasks))
	}
}

// TestGetAll tests retrieving all tasks
func TestGetAll(t *testing.T) {
	h := setupHandler()

	// Create HTTP request and response recorder
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	w := httptest.NewRecorder()

	// Call the handler
	h.GetAll(w, req)

	// Check status code
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	// Parse response
	var response tasks.ListTasksResponse
	if err := protojson.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify tasks
	if len(response.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(response.Tasks))
	}

	expectedID := "550e8400-e29b-41d4-a716-446655440000"
	if response.Tasks[0].Id != expectedID {
		t.Errorf("expected task ID '%s', got '%s'", expectedID, response.Tasks[0].Id)
	}
}

// TestGetAllEmpty tests getting all tasks when empty
func TestGetAllEmpty(t *testing.T) {
	h := NewTaskHandler()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	w := httptest.NewRecorder()

	h.GetAll(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response tasks.ListTasksResponse
	if err := protojson.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(response.Tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(response.Tasks))
	}
}

// TestCreate tests creating a new task
func TestCreate(t *testing.T) {
	h := NewTaskHandler()

	// Create request body
	reqBody := &tasks.CreateTaskRequest{
		Title:       "New Task",
		Description: "New Description",
	}

	bodyBytes, err := protojson.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	h.Create(w, req)

	// Check status code
	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	// Parse response
	var response tasks.GetTaskResponse
	if err := protojson.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify task fields
	task := response.Task
	if task.Title != "New Task" {
		t.Errorf("expected title 'New Task', got '%s'", task.Title)
	}

	if task.Description != "New Description" {
		t.Errorf("expected description 'New Description', got '%s'", task.Description)
	}

	if task.Completed {
		t.Error("expected completed to be false")
	}

	if task.Id == "" {
		t.Error("expected non-empty ID")
	}

	// Verify task was added to map
	taskID, err := uuid.Parse(task.Id)
	if err != nil {
		t.Fatalf("failed to parse task ID: %v", err)
	}

	h.mu.RLock()
	_, exists := h.tasks[taskID]
	h.mu.RUnlock()

	if !exists {
		t.Error("task was not added to map")
	}
}

// TestCreateValidation tests validation errors
func TestCreateValidation(t *testing.T) {
	h := NewTaskHandler()

	tests := []struct {
		name        string
		title       string
		description string
		wantStatus  int
	}{
		{
			name:        "empty title",
			title:       "",
			description: "Valid description",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "title too long",
			title:       string(make([]byte, 101)), // 101 chars
			description: "Valid description",
			wantStatus:  http.StatusBadRequest,
		},
		{
			name:        "description too long",
			title:       "Valid title",
			description: string(make([]byte, 501)), // 501 chars
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody := &tasks.CreateTaskRequest{
				Title:       tt.title,
				Description: tt.description,
			}

			bodyBytes, _ := protojson.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			h.Create(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// TestCreateInvalidJSON tests invalid JSON handling
func TestCreateInvalidJSON(t *testing.T) {
	h := NewTaskHandler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

// TestConcurrentAccess tests thread safety
func TestConcurrentAccess(t *testing.T) {
	h := NewTaskHandler()

	// Spawn multiple goroutines to test concurrent access
	const numGoroutines = 100
	done := make(chan bool)

	// Create tasks concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			reqBody := &tasks.CreateTaskRequest{
				Title:       "Concurrent Task",
				Description: "Test",
			}

			bodyBytes, _ := protojson.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(bodyBytes))
			w := httptest.NewRecorder()

			h.Create(w, req)
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all tasks were created
	h.mu.RLock()
	taskCount := len(h.tasks)
	h.mu.RUnlock()

	if taskCount != numGoroutines {
		t.Errorf("expected %d tasks, got %d (possible race condition)", numGoroutines, taskCount)
	}
}

// Benchmark for Create operation
func BenchmarkCreate(b *testing.B) {
	h := NewTaskHandler()

	reqBody := &tasks.CreateTaskRequest{
		Title:       "Benchmark Task",
		Description: "Benchmark Description",
	}

	bodyBytes, _ := protojson.Marshal(reqBody)

	b.ResetTimer() // Don't count setup time

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()
		h.Create(w, req)
	}
}

// Benchmark for GetAll operation
func BenchmarkGetAll(b *testing.B) {
	h := setupHandler()

	// Add more tasks for realistic benchmark
	for i := 0; i < 100; i++ {
		now := timestamppb.Now()
		taskID := uuid.New()
		h.tasks[taskID] = &tasks.Task{
			Id:          taskID.String(),
			Title:       "Task",
			Description: "Description",
			Completed:   false,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
		w := httptest.NewRecorder()
		h.GetAll(w, req)
	}
}
