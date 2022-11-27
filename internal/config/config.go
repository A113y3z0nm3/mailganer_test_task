package config

import (
	"fmt"
	"mailganer_test_task/internal/handlers"
	"mailganer_test_task/internal/services"
	email "mailganer_test_task/internal/transport"
	log "mailganer_test_task/pkg/logger"

	"github.com/caarlos0/env"
)

// Config Структура с конфигурацией приложения
type Config struct {
	Log		*log.ConfigLog
	Service	*services.MailingServiceConfig
	Email	*email.EmailConfig
	Message	*email.Message
	Handler	*handlers.MailingHandlerConfig
}

// LoadConfig загружает конфигурацию из env
func LoadConfig() (*Config, error) {
	// Инициализация конфигурации,
	// если был добавлен новый конфигу куда-либо, то
	// необходимо проинициализировать его тут, иначе будет nil pointer
	config := &Config{
		Log:		&log.ConfigLog{},
		Service:	&services.MailingServiceConfig{},
		Email:		&email.EmailConfig{},
		Message:	&email.Message{},
		Handler:	&handlers.MailingHandlerConfig{},
	}

	if err := env.Parse(config); err != nil {
		return nil, fmt.Errorf("unable to load configuration. Error: %s", err)
	}

	return config, nil
}
