package configuration

import (
	xlogger "github.com/syth0le/gopnik/logger"
	xservers "github.com/syth0le/gopnik/servers"

	"time"
)

type Config struct {
	Logger             xlogger.LoggerConfig  `yaml:"logger"`
	Application        ApplicationConfig     `yaml:"application"`
	PublicServer       xservers.ServerConfig `yaml:"public_server"`
	AdminServer        xservers.ServerConfig `yaml:"admin_server"`
	ConnectionsStorage RedisConfig           `yaml:"redis"`
	Queue              RabbitConfig          `yaml:"queue"`
}

func (c *Config) Validate() error {
	return nil // todo
}

type ApplicationConfig struct {
	GracefulShutdownTimeout time.Duration `yaml:"graceful_shutdown_timeout"`
	ForceShutdownTimeout    time.Duration `yaml:"force_shutdown_timeout"`
	App                     string        `yaml:"app"`
}

func (c *ApplicationConfig) Validate() error {
	return nil // todo
}

type RedisConfig struct {
	Enable             bool          `yaml:"enable"`
	Address            string        `yaml:"address"`
	Password           string        `yaml:"password" env:"CACHE_DB_PASSWORD"`
	Database           int           `yaml:"database"`
	ExpirationDuration time.Duration `yaml:"expiration_duration"`
	HeaterDuration     time.Duration `yaml:"heater_duration"`
	MaxListRange       int64         `yaml:"max_list_range"`
}

type RabbitConfig struct {
	Enable       bool   `yaml:"enable"`
	Address      string `yaml:"address"`
	QueueName    string `yaml:"queue_name"`
	ExchangeName string `yaml:"exchange_name"`
}
