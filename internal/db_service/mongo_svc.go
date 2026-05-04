package db_service

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbService[DocType interface{}] interface {
	CreateDocument(ctx context.Context, id string, document *DocType) error
	FindDocument(ctx context.Context, id string) (*DocType, error)
	FindAll(ctx context.Context) ([]DocType, error)
	UpdateDocument(ctx context.Context, id string, document *DocType) error
	DeleteDocument(ctx context.Context, id string) error
	Disconnect(ctx context.Context) error
}

var ErrNotFound = fmt.Errorf("document not found")
var ErrConflict = fmt.Errorf("conflict: document already exists")

type MongoServiceConfig struct {
	ServerHost string
	ServerPort int
	UserName   string
	Password   string
	DbName     string
	Collection string
	Timeout    time.Duration
}

type mongoSvc[DocType interface{}] struct {
	MongoServiceConfig
	client     atomic.Pointer[mongo.Client]
	clientLock sync.Mutex
}

func NewMongoService[DocType interface{}](config MongoServiceConfig) DbService[DocType] {
	enviro := func(name string, defaultValue string) string {
		if value, ok := os.LookupEnv(name); ok {
			return value
		}
		return defaultValue
	}

	svc := &mongoSvc[DocType]{}
	svc.MongoServiceConfig = config

	if svc.ServerHost == "" {
		svc.ServerHost = enviro("OR_PLANNER_API_MONGODB_HOST", "localhost")
	}
	if svc.ServerPort == 0 {
		port := enviro("OR_PLANNER_API_MONGODB_PORT", "27017")
		if p, err := strconv.Atoi(port); err == nil {
			svc.ServerPort = p
		} else {
			svc.ServerPort = 27017
		}
	}
	if svc.UserName == "" {
		svc.UserName = enviro("OR_PLANNER_API_MONGODB_USERNAME", "")
	}
	if svc.Password == "" {
		svc.Password = enviro("OR_PLANNER_API_MONGODB_PASSWORD", "")
	}
	if svc.DbName == "" {
		svc.DbName = enviro("OR_PLANNER_API_MONGODB_DATABASE", "orp-or-planner")
	}
	if svc.Collection == "" {
		svc.Collection = enviro("OR_PLANNER_API_MONGODB_COLLECTION", "default")
	}
	if svc.Timeout == 0 {
		seconds := enviro("OR_PLANNER_API_MONGODB_TIMEOUT_SECONDS", "10")
		if s, err := strconv.Atoi(seconds); err == nil {
			svc.Timeout = time.Duration(s) * time.Second
		} else {
			svc.Timeout = 10 * time.Second
		}
	}

	log.Printf("MongoDB config: //%v@%v:%v/%v/%v",
		svc.UserName, svc.ServerHost, svc.ServerPort, svc.DbName, svc.Collection)
	return svc
}

func (m *mongoSvc[DocType]) connect(ctx context.Context) (*mongo.Client, error) {
	client := m.client.Load()
	if client != nil {
		return client, nil
	}

	m.clientLock.Lock()
	defer m.clientLock.Unlock()
	client = m.client.Load()
	if client != nil {
		return client, nil
	}

	ctx, cancel := context.WithTimeout(ctx, m.Timeout)
	defer cancel()

	uri := fmt.Sprintf("mongodb://%v:%v", m.ServerHost, m.ServerPort)
	if m.UserName != "" {
		uri = fmt.Sprintf("mongodb://%v:%v@%v:%v", m.UserName, m.Password, m.ServerHost, m.ServerPort)
	}

	c, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetConnectTimeout(10*time.Second))
	if err != nil {
		return nil, err
	}
	m.client.Store(c)
	return c, nil
}

func (m *mongoSvc[DocType]) Disconnect(ctx context.Context) error {
	client := m.client.Load()
	if client == nil {
		return nil
	}
	m.clientLock.Lock()
	defer m.clientLock.Unlock()
	client = m.client.Load()
	defer m.client.Store(nil)
	if client != nil {
		return client.Disconnect(ctx)
	}
	return nil
}

func (m *mongoSvc[DocType]) collection(ctx context.Context) (*mongo.Collection, error) {
	c, err := m.connect(ctx)
	if err != nil {
		return nil, err
	}
	return c.Database(m.DbName).Collection(m.Collection), nil
}

func (m *mongoSvc[DocType]) CreateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, cancel := context.WithTimeout(ctx, m.Timeout)
	defer cancel()
	col, err := m.collection(ctx)
	if err != nil {
		return err
	}
	r := col.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch r.Err() {
	case nil:
		return ErrConflict
	case mongo.ErrNoDocuments:
	default:
		return r.Err()
	}
	_, err = col.InsertOne(ctx, document)
	return err
}

func (m *mongoSvc[DocType]) FindDocument(ctx context.Context, id string) (*DocType, error) {
	ctx, cancel := context.WithTimeout(ctx, m.Timeout)
	defer cancel()
	col, err := m.collection(ctx)
	if err != nil {
		return nil, err
	}
	r := col.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch r.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return nil, ErrNotFound
	default:
		return nil, r.Err()
	}
	var doc *DocType
	if err := r.Decode(&doc); err != nil {
		return nil, err
	}
	return doc, nil
}

func (m *mongoSvc[DocType]) FindAll(ctx context.Context) ([]DocType, error) {
	ctx, cancel := context.WithTimeout(ctx, m.Timeout)
	defer cancel()
	col, err := m.collection(ctx)
	if err != nil {
		return nil, err
	}
	cur, err := col.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	results := []DocType{}
	for cur.Next(ctx) {
		var doc DocType
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		results = append(results, doc)
	}
	return results, nil
}

func (m *mongoSvc[DocType]) UpdateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, cancel := context.WithTimeout(ctx, m.Timeout)
	defer cancel()
	col, err := m.collection(ctx)
	if err != nil {
		return err
	}
	r := col.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch r.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return ErrNotFound
	default:
		return r.Err()
	}
	_, err = col.ReplaceOne(ctx, bson.D{{Key: "id", Value: id}}, document)
	return err
}

func (m *mongoSvc[DocType]) DeleteDocument(ctx context.Context, id string) error {
	ctx, cancel := context.WithTimeout(ctx, m.Timeout)
	defer cancel()
	col, err := m.collection(ctx)
	if err != nil {
		return err
	}
	r := col.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch r.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return ErrNotFound
	default:
		return r.Err()
	}
	_, err = col.DeleteOne(ctx, bson.D{{Key: "id", Value: id}})
	return err
}
