package models

import (
	"bytes"
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
	Template	*bytes.Buffer
}

// TemplateData Структура для заполнения шаблона
type TemplateData struct {
	Firstname	string
	Lastname	string
	URL			string
}
