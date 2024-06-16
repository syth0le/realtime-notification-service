package configuration

import (
	"time"

	xclients "github.com/syth0le/gopnik/clients"
	xlogger "github.com/syth0le/gopnik/logger"
	xservers "github.com/syth0le/gopnik/servers"
)

const (
	defaultAppName = "realtime-notification"
)

func NewDefaultConfig() *Config {
	return &Config{
		Logger: xlogger.LoggerConfig{
			Level:       xlogger.InfoLevel,
			Encoding:    "console",
			Path:        "stdout",
			Environment: xlogger.Development,
		},
		Application: ApplicationConfig{
			GracefulShutdownTimeout: 15 * time.Second,
			ForceShutdownTimeout:    20 * time.Second,
			App:                     defaultAppName,
		},
		PublicServer: xservers.ServerConfig{
			Enable:   false,
			Endpoint: "",
			Port:     0,
		},
		AdminServer: xservers.ServerConfig{
			Enable:   false,
			Endpoint: "",
			Port:     0,
		},
		Queue: RabbitConfig{
			Enable:       false,
			Address:      "",
			QueueName:    "",
			ExchangeName: "",
		},
		AuthClient: AuthClientConfig{
			Enable: false,
			Conn: xclients.GRPCClientConnConfig{
				Endpoint:              "",
				UserAgent:             defaultAppName,
				MaxRetries:            0,
				TimeoutBetweenRetries: 0,
				InitTimeout:           0,
				EnableCompressor:      false,
			},
		},
	}
}
