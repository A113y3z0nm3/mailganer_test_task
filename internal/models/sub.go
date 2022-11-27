package models

import (
	"bytes"
	"context"
	"html/template"
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// Sub Структура с информацией о подписчике
type Sub struct {
	BirthDay	time.Time
	Firstname	string
	Lastname	string
	Email		string
	CronID		cron.EntryID
	UUID		uuid.UUID
	Uid			string
	Template	*bytes.Buffer
}

// BuildTemplate Собирает шаблон письма для подписчика и записывает его в буфер
func (sub *Sub) BuildTemplate(ctx context.Context, templatesPath string) error {
	// Парсим html файл в шаблон
	tmplt, err := template.ParseFiles(templatesPath)
	if err != nil {
		return err
	}

	// Создаем байтовый буфер
	buf := new(bytes.Buffer)

	// Записываем данные подписчика в шаблон, а шаблон в буфер
	if err := tmplt.Execute(buf, sub); err != nil {
		return err
	}

	sub.Template = buf

	return nil
}
