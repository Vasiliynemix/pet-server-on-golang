package main

import (
	"PetProjectGo/internal/config"
	"PetProjectGo/internal/server"
	"PetProjectGo/pkg/logging"
	"PetProjectGo/pkg/storage/mongodb"
	"PetProjectGo/pkg/storage/postgres"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"time"
)

func main() {
	cfg, err := config.InitConfiguration("./config")
	if err != nil {
		panic(err)
	}
	logger := logging.NewLogger(
		logging.InitLogger(
			cfg.Log.StructDateFormat, cfg.Log.PathInfo, cfg.Log.PathDebug, cfg.Log.LogLevel,
		),
	)
	logger.Debug("Debug mode")
	logger.Debug("config", zap.Any("config", cfg))
	go logging.ClearLogFiles(cfg.Log.PathInfo, cfg.Log.PathDebug, logger)

	mongoDB := mongodb.NewMongoDB(logger, &cfg.Mongo)
	defer mongoDB.Disconnect()

	err = mongoDB.Connect()
	if err != nil {
		logger.Fatal("MongoDB connection failed", zap.Error(err))
	}

	for {
		if mongoDB.IsConnected() {
			logger.Info(
				"MongoDB connection success",
				zap.String("host", cfg.Mongo.Host),
				zap.String("database", cfg.Mongo.Database),
				zap.Int("port", cfg.Mongo.Port),
			)
			break
		}
		time.Sleep(1 * time.Second)
	}

	postgresDB, err := postgres.NewPostgresConnection(logger, &cfg.Postgres)
	if err != nil {
		logger.Fatal("Error connecting to postgresRepo", zap.Error(err))
	}

	logger.Info(
		"Postgres connection success",
		zap.String("host", cfg.Postgres.Host),
		zap.String("database", cfg.Postgres.Database),
		zap.Int("port", cfg.Postgres.Port),
	)

	err = postgres.Migrations(postgresDB, logger)
	if err != nil {
		logger.Error("Postgres migration error", zap.Error(err))
	}

	srv := server.NewWebServer(logger, cfg, mongoDB, postgresDB)
	go srv.Run()

	//Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	takeSig := <-sigChan
	logger.Info("Shutdown signal", zap.String("signal", takeSig.String()))
}
