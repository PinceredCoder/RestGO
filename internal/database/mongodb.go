package database

import (
	"context"
	"fmt"
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
}

func NewMongoDatabase(ctx context.Context, uri, dbName string) (*MongoDatabase, error) {
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
	}

	return &MongoDatabase{
		client:   client,
		database: database,
		taskRepo: taskRepo,
	}, nil
}

func (m *MongoDatabase) Ping(ctx context.Context) error {
	return m.client.Ping(ctx, nil)
}

// Disconnect closes the database connection
func (m *MongoDatabase) Disconnect(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// GetTaskRepository returns the task repository
func (m *MongoDatabase) GetTaskRepository() TaskRepository {
	return m.taskRepo
}

// MongoTaskRepository implements the TaskRepository interface
type MongoTaskRepository struct {
	collection *mongo.Collection
}

// Create inserts a new task
func (r *MongoTaskRepository) Create(ctx context.Context, task *Task) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, task)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}

	return nil
}

// FindByID retrieves a task by ID
func (r *MongoTaskRepository) FindByID(ctx context.Context, id uuid.UUID) (*Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var task Task
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&task)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // Not found, return nil without error
		}
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	return &task, nil
}

// FindAll retrieves all tasks
func (r *MongoTaskRepository) FindAll(ctx context.Context) ([]*Task, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find tasks: %w", err)
	}
	defer cursor.Close(ctx)

	var tasks []*Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, fmt.Errorf("failed to decode tasks: %w", err)
	}

	return tasks, nil
}

// Update modifies an existing task
func (r *MongoTaskRepository) Update(ctx context.Context, id uuid.UUID, task *Task) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"title":       task.Title,
			"description": task.Description,
			"completed":   task.Completed,
			"updatedAt":   task.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	if result.MatchedCount == 0 {
		return nil // Not found, handled by caller
	}

	return nil
}

// Delete removes a task by ID
func (r *MongoTaskRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	if result.DeletedCount == 0 {
		return nil // Not found, handled by caller
	}

	return nil
}
