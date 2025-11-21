package database

import (
	"context"

	"github.com/google/uuid"
)

type Database interface {
	Ping(ctx context.Context) error
	Disconnect(ctx context.Context) error
	GetTaskRepository() TaskRepository
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	FindByID(ctx context.Context, id uuid.UUID) (*Task, error)
	FindAll(ctx context.Context) ([]*Task, error)
	Update(ctx context.Context, id uuid.UUID, task *Task) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type Task struct {
	ID          uuid.UUID `bson:"_id"`
	Title       string    `bson:"title"`
	Description string    `bson:"description"`
	Completed   bool      `bson:"completed"`
	CreatedAt   int64     `bson:"createdAt"`
	UpdatedAt   int64     `bson:"updatedAt"`
}
