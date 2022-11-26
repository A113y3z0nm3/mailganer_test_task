package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
)

// Sub
type Sub struct {
	BirthDay	time.Time
	Firstname	string
	Lastname	string
	Email		string
	CronID		cron.EntryID
	UUID		uuid.UUID
}
