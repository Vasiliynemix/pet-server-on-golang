package mongodb

import (
	"PetProjectGo/internal/config"
	"PetProjectGo/pkg/logging"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/net/context"
)

type MongoDB struct {
	client    *mongo.Client
	log       *logging.Logger
	cfg       *config.MongoDBConnectionConfig
	connected bool
}

func NewMongoDB(log *logging.Logger, cfg *config.MongoDBConnectionConfig) *MongoDB {
	return &MongoDB{
		client:    nil,
		log:       log,
		cfg:       cfg,
		connected: false,
	}
}

func (m *MongoDB) Connect() error {
	var err error
	uri := buildUri(m.cfg)

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)
	m.client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		return fmt.Errorf("MongoDB connection error: %w", err)
	}

	var result bson.M
	if err = m.client.Database("admin").RunCommand(context.TODO(), bson.M{"ping": 1}).Decode(&result); err != nil {
		return fmt.Errorf("MongoDB connection error: %w", err)
	}

	m.connected = true
	return nil
}

func (m *MongoDB) IsConnected() bool {
	return m.connected
}

func (m *MongoDB) Disconnect() {
	if m.client != nil {
		m.client.Disconnect(context.TODO())
	}
}

func (m *MongoDB) GetDB() *mongo.Database {
	return m.client.Database(m.cfg.Database)
}

func (m *MongoDB) GetCollection(name string) *mongo.Collection {
	return m.GetDB().Collection(name)
}

func buildUri(cfg *config.MongoDBConnectionConfig) string {
	if cfg.Username != "" && cfg.Password != "" {
		return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin",
			cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	} else {
		return fmt.Sprintf("mongodb://%s:%d/%s?authSource=admin",
			cfg.Host, cfg.Port, cfg.Database)
	}
}
