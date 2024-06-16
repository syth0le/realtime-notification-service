package configuration

import (
	xclients "github.com/syth0le/gopnik/clients"
	xlogger "github.com/syth0le/gopnik/logger"
	xservers "github.com/syth0le/gopnik/servers"

	"time"
)

type Config struct {
	Logger       xlogger.LoggerConfig  `yaml:"logger"`
	Application  ApplicationConfig     `yaml:"application"`
	PublicServer xservers.ServerConfig `yaml:"public_server"`
	AdminServer  xservers.ServerConfig `yaml:"admin_server"`
	Queue        RabbitConfig          `yaml:"queue"`
	AuthClient   AuthClientConfig      `yaml:"auth"`
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

type RabbitConfig struct {
	Enable       bool   `yaml:"enable"`
	Address      string `yaml:"address"`
	QueueName    string `yaml:"queue_name"`
	ExchangeName string `yaml:"exchange_name"`
}

func (c *RabbitConfig) Validate() error {
	if !c.Enable {
		return nil
	}

	return nil // todo
}

type AuthClientConfig struct {
	Enable bool                          `yaml:"enable"`
	Conn   xclients.GRPCClientConnConfig `yaml:"conn"`
}

func (c *AuthClientConfig) Validate() error {
	if !c.Enable {
		return nil
	}

	return nil // todo
}
