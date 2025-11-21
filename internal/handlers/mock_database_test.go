package handlers

import (
	"context"
	"sync"

	"github.com/PinceredCoder/restGo/internal/database"
	"github.com/google/uuid"
)

// MockDatabase implements the database.Database interface for testing
type MockDatabase struct {
	taskRepo *MockTaskRepository
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{
		taskRepo: &MockTaskRepository{
			tasks: make(map[uuid.UUID]*database.Task),
		},
	}
}

func (m *MockDatabase) Ping(ctx context.Context) error {
	return nil
}

func (m *MockDatabase) Disconnect(ctx context.Context) error {
	return nil
}

func (m *MockDatabase) GetTaskRepository() database.TaskRepository {
	return m.taskRepo
}

// MockTaskRepository implements the database.TaskRepository interface for testing
type MockTaskRepository struct {
	mu    sync.RWMutex
	tasks map[uuid.UUID]*database.Task
}

func (r *MockTaskRepository) Create(ctx context.Context, task *database.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return nil
}

func (r *MockTaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*database.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	task, exists := r.tasks[id]
	if !exists {
		return nil, nil
	}
	return task, nil
}

func (r *MockTaskRepository) FindAll(ctx context.Context) ([]*database.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]*database.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (r *MockTaskRepository) Update(ctx context.Context, id uuid.UUID, task *database.Task) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.tasks[id]
	if !exists {
		return nil // Mimics MongoDB behavior
	}

	r.tasks[id] = task
	return nil
}

func (r *MockTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.tasks[id]
	if !exists {
		return nil // Mimics MongoDB behavior
	}

	delete(r.tasks, id)
	return nil
}
