package mocks

import (
	"time"
	"context"

	"github.com/stretchr/testify/mock"
	cron "github.com/robfig/cron/v3"
)

// MockSched mock реализация планировщика
type MockSched struct {
	mock.Mock
}

// AddFunc mock метода
func (m *MockSched)AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	arg := m.Called(spec, cmd)

	var arg1 cron.EntryID

	if arg.Get(0) != nil {
		arg1 = arg.Get(0).(cron.EntryID)
	}

	var arg2 error

	if arg.Get(1) != nil {
		arg2 = arg.Get(1).(error)
	}

	return arg1, arg2
}

// AddJob mock метода
func (m *MockSched)AddJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	arg := m.Called(spec, cmd)

	var arg1 cron.EntryID

	if arg.Get(0) != nil {
		arg1 = arg.Get(0).(cron.EntryID)
	}

	var arg2 error

	if arg.Get(1) != nil {
		arg2 = arg.Get(1).(error)
	}

	return arg1, arg2
}

// Entries mock метода
func (m *MockSched)Entries() []cron.Entry {
	arg := m.Called()

	var arg1 []cron.Entry

	if arg.Get(0) != nil {
		arg1 = arg.Get(0).([]cron.Entry)
	}

	return arg1
}

// Entry mock метода
func (m *MockSched)Entry(id cron.EntryID) cron.Entry {
	arg := m.Called(id)

	var arg1 cron.Entry

	if arg.Get(0) != nil {
		arg1 = arg.Get(0).(cron.Entry)
	}

	return arg1
}

// Location mock метода
func (m *MockSched)Location() *time.Location {
	arg := m.Called()

	var arg1 *time.Location

	if arg.Get(0) != nil {
		arg1 = arg.Get(0).(*time.Location)
	}

	return arg1
}

// Remove mock метода
func (m *MockSched)Remove(id cron.EntryID) {
	m.Called(id)
}

// Run mock метода
func (m *MockSched)Run() {
}

// Schedule mock метода
func (m *MockSched)Schedule(schedule cron.Schedule, cmd cron.Job) cron.EntryID {
	arg := m.Called(schedule, cmd)

	var arg1 cron.EntryID

	if arg.Get(0) != nil {
		arg1 = arg.Get(0).(cron.EntryID)
	}

	return arg1
}

// Start mock метода
func (m *MockSched)Start() {
	
}

// Stop mock метода
func (m *MockSched)Stop() context.Context {
	arg := m.Called()

	var arg1 context.Context

	if arg.Get(0) != nil {
		arg1 = arg.Get(0).(context.Context)
	}

	return arg1
}
