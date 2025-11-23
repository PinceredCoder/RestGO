package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDatabase struct {
	client   *mongo.Client
	database *mongo.Database
	taskRepo *MongoTaskRepository
	logger   *slog.Logger
}

func NewMongoDatabase(ctx context.Context, uri, dbName string) (*MongoDatabase, error) {
	logger := slog.Default()

	clientOptions := options.Client().ApplyURI(uri)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(dbName)

	taskRepo := &MongoTaskRepository{
		collection: database.Collection("tasks"),
		logger:     logger,
	}

	return &MongoDatabase{
		client:   client,
		database: database,
		taskRepo: taskRepo,
		logger:   logger,
	}, nil
}

func (m *MongoDatabase) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

func (m *MongoDatabase) Disconnect(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

func (m *MongoDatabase) GetTaskRepository() TaskRepository {
	return m.taskRepo
}

type MongoTaskRepository struct {
	collection *mongo.Collection
	logger     *slog.Logger
}

func (r *MongoTaskRepository) Create(ctx context.Context, task *Task) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("Creating task in MongoDB", "task_id", task.ID)

	_, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		r.logger.Error("MongoDB insert failed", "error", err, "task_id", task.ID)
		return fmt.Errorf("failed to create task: %w", err)
	}

	r.logger.Debug("Task created in MongoDB", "task_id", task.ID)
	return nil
}

func (r *MongoTaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("Finding task by ID in MongoDB", "task_id", id)

	var task Task
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Debug("Task not found in MongoDB", "task_id", id)
			return nil, nil
		}
		r.logger.Error("MongoDB find failed", "error", err, "task_id", id)
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	r.logger.Debug("Task found in MongoDB", "task_id", id)
	return &task, nil
}

func (r *MongoTaskRepository) FindAll(ctx context.Context) ([]*Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("Finding all tasks in MongoDB")

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		r.logger.Error("MongoDB find all failed", "error", err)
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []*Task
	if err := cursor.All(ctx, &tasks); err != nil {
		r.logger.Error("MongoDB decode failed", "error", err)
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	r.logger.Debug("All tasks retrieved from MongoDB", "count", len(tasks))
	return tasks, nil
}

func (r *MongoTaskRepository) Update(ctx context.Context, id uuid.UUID, task *Task) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("Updating task in MongoDB", "task_id", id)

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"title":       task.Title,
			"description": task.Description,
			"completed":   task.Completed,
			"updatedAt":   task.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("MongoDB update failed", "error", err, "task_id", id)
		return fmt.Errorf("failed to update task: %w", err)
	}

	r.logger.Debug("Task updated in MongoDB", "task_id", id)
	return nil
}

func (r *MongoTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	r.logger.Debug("Deleting task from MongoDB", "task_id", id)

	filter := bson.M{"_id": id}
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("MongoDB delete failed", "error", err, "task_id", id)
		return fmt.Errorf("failed to delete task: %w", err)
	}

	r.logger.Debug("Task deleted from MongoDB", "task_id", id)
	return nil
}
