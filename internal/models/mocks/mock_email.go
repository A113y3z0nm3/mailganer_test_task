package mocks

import (
	"mailganer_test_task/internal/transport"

	"github.com/stretchr/testify/mock"
)

// MockClient mock реализация email клиента
type MockClient struct {
	mock.Mock
}

// Send mock метода
func (m *MockClient) Send(msg *email.Message) error {
	arg := m.Called(msg)

	var arg1 error

	if arg.Get(0) != nil {
		arg1 = arg.Get(0).(error)
	}

	return arg1
}
