package models

// ConfigLog конфигурация для логгера
type ConfigLog struct {
	Mode	string	`env:"LOG_MODE"`
	Level	string	`env:"LOG_LEVEL"`
	Output	string	`env:"LOG_OUTPUT"`
}