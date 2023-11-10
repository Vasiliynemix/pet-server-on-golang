package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Log      Logger                   `mapstructure:"log"`
	Web      ServerConfig             `mapstructure:"web"`
	App      AppConfig                `mapstructure:"app"`
	Postgres PostgresConnectionConfig `mapstructure:"postgresRepo"`
	Mongo    MongoDBConnectionConfig  `mapstructure:"mongoRepo"`
}

type AppConfig struct {
	SecretKeyToken                    string        `mapstructure:"secret_key_token"`
	TokenExpirationTimeMinutes        time.Duration `mapstructure:"token_expiration_time_minutes"`
	RefreshTokenExpirationTimeMinutes time.Duration `mapstructure:"refresh_token_expiration_time_minutes"`
	PasswordMinLength                 int           `mapstructure:"password_min_length"`
}

type Logger struct {
	PathInfo         string `mapstructure:"path_info"`
	PathDebug        string `mapstructure:"path_debug"`
	LogLevel         string `mapstructure:"log_level"`
	StructDateFormat string `mapstructure:"struct_date"`
}

type ServerConfig struct {
	Address     string        `mapstructure:"address"`
	Timeout     time.Duration `mapstructure:"timeout"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
}

type MongoDBConnectionConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password" json:"-"`
}

type PostgresConnectionConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password" json:"-"`
	Pool     Pool   `mapstructure:"pool"`
}

type Pool struct {
	MaxIdleConnections int           `mapstructure:"max_idle_connections"`
	MaxOpenConnections int           `mapstructure:"max_open_connections"`
	IdleTimeout        time.Duration `mapstructure:"idle_timeout"`
}

func InitConfiguration(pathToConfig string) (*Config, error) {
	var cfg *Config

	loadDefault()
	viper.AutomaticEnv()
	loadFile(pathToConfig)

	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func loadDefault() {
	viper.SetDefault("log.path_info", "./logs/info.log")
	viper.SetDefault("log.path_debug", "./logs/debug.log")
	viper.SetDefault("log.log_level", "debug")
	viper.SetDefault("log.struct_date", "02-01-2006 15:04:05")

	viper.SetDefault("web.address", "127.0.0.1:8080")
	viper.SetDefault("web.timeout", 10*time.Second)
	viper.SetDefault("web.idle_timeout", 60*time.Second)

	viper.SetDefault("app.password_min_length", 8)
	viper.SetDefault("app.secret_key_token", "secret_key_token")
	viper.SetDefault("app.token_expiration_time_minutes", 5)
	viper.SetDefault("app.refresh_token_expiration_time_minutes", 60*24)

	viper.SetDefault("mongoRepo.host", "localhost")
	viper.SetDefault("mongoRepo.port", 27017)
	viper.SetDefault("mongoRepo.database", "auth_mongo_db")
	viper.SetDefault("mongoRepo.username", "")
	viper.SetDefault("mongoRepo.password", "")

	viper.SetDefault("postgresRepo.host", "localhost")
	viper.SetDefault("postgresRepo.port", 5432)
	viper.SetDefault("postgresRepo.database", "auth_postgres_db")
	viper.SetDefault("postgresRepo.username", "user")
	viper.SetDefault("postgresRepo.password", "password")

	viper.SetDefault("postgresRepo.pool.max_idle_connections", 1)
	viper.SetDefault("postgresRepo.pool.max_open_connections", 10)
	viper.SetDefault("postgresRepo.pool.idle_timeout", 300*time.Second)
}

func loadFile(pathToConfig string) {
	viper.SetConfigName("dev")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(pathToConfig)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(fmt.Sprintf("Error reading config file: %s", err))
	}
}
