package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"mailganer_test_task/internal/models"
	"mailganer_test_task/internal/models/mocks"
	email "mailganer_test_task/internal/transport"
	log "mailganer_test_task/pkg/logger"

	"github.com/google/uuid"
	cron "github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type SearchSubTestCase struct {
	CaseName	string
	Sub			models.Sub
}

type BuildTemplateTestCase struct {
	CaseName	string
	Sub			models.Sub
	Path		string
}

type SendEmailTestCase struct {
	CaseName	string
	Sub			models.Sub
	Err			error
}

type PushEmailTestCase struct {
	CaseName	string
	Sub			models.Sub
	Path		string
	Err			error
}

type AddSubTemplateTestCase struct {
	CaseName	string
	Sub			models.Sub
	Path		string
	ErrSched	error
}

type RemoveSubTestCase struct {
	CaseName	string
	Sub			models.Sub
}

func TestMailingService(t *testing.T) {
	a := assert.New(t)

	client := new(mocks.MockClient)
	sched := new(mocks.MockSched)
	logger, err := log.InitLogger(log.DevConfig)
	if err != nil {
		logger.Fatal("unable to init logger")
	}

	message := &email.Message{}

	service := NewMailingService(&MailingServiceConfig{
		EmailClient:	client,
		TemplatesPath:	"../templates/message.html",
		Host:			"0.0.0.0",
		Port:			"8080",
		Logger:			logger,
		Message:		message,
	})

	TestUid := uuid.New()
	service.subs[TestUid] = models.Sub{
		UUID: TestUid,
	}
	service.sched = sched

	client.ExpectedCalls	= []*mock.Call{}
	sched.ExpectedCalls		= []*mock.Call{}

	t.Run("WriteOpening", func(t *testing.T){
		t.Log("WriteOpening: Begin")

		ctx := context.TODO()
		service.WriteOpening(ctx, TestUid)

		t.Log("WriteOpening: Done")
	})

	t.Run("AddSubTemplate", func(t *testing.T) {
		

		testCases := []AddSubTemplateTestCase{	
			{
				CaseName:	"Already registered",
				Sub:		models.Sub{
					UUID: TestUid,
				},
				Path:		service.templatesPath,
			},
			{
				CaseName:	"Invalid path",
				Sub:		models.Sub{},
				Path:		"",
			},
			{
				CaseName:	"Err in email sched (test)",
				Sub:		models.Sub{},
				Path:		service.templatesPath,
				ErrSched:	errors.New("test error"),
			},
			{
				CaseName:	"Success",
				Sub:		models.Sub{
					BirthDay: time.Date(2006, 1, 1, 1, 1, 1, 1, time.UTC),
				},
				Path:		service.templatesPath,
			},
		}

		for _, c := range testCases {
			t.Logf("%s: Begin", c.CaseName)

			if (c.CaseName == "Success") || (c.CaseName == "Err in email sched (test)") {
				sched.On("AddFunc", formatSched(c.Sub.BirthDay), mock.AnythingOfType("func()")).Return(cron.EntryID(1), c.ErrSched)
			}

			service.templatesPath = c.Path
			ctx := context.TODO()

			err := service.AddSubTemplate(ctx, c.Sub)

			a.NoError(err, err)

			if (c.CaseName == "Success") || (c.CaseName == "Err in email sched (test)") {
				sched.AssertCalled(t, "AddFunc", formatSched(c.Sub.BirthDay), mock.AnythingOfType("func()"))
			}

			t.Logf("%s: Done", c.CaseName)

			if (c.CaseName == "Success") || (c.CaseName == "Err in email sched (test)") {
				sched.ExpectedCalls = []*mock.Call{}
			}
		}
	})

	t.Run("RemoveSub", func(t *testing.T) {
		testCases := []RemoveSubTestCase{
			{
				CaseName:	"Success",
				Sub:		models.Sub{
					UUID: TestUid,
				},
			},
			{
				CaseName:	"Not found",
				Sub:		models.Sub{},
			},
		}

		for _, c := range testCases {
			t.Logf("%s: Begin", c.CaseName)

			if c.CaseName == "Success" {
				sched.On("Remove", mock.AnythingOfType("cron.EntryID"))
			}

			ctx := context.TODO()

			err := service.RemoveSub(ctx, c.Sub)

			a.NoError(err, err)

			if c.CaseName == "Success" {
				sched.AssertCalled(t, "Remove", mock.AnythingOfType("cron.EntryID"))
			}

			t.Logf("%s: Done", c.CaseName)

			if c.CaseName == "Success" {
				sched.ExpectedCalls = []*mock.Call{}
			}
		}
	})

	t.Run("PushEmail", func(t *testing.T) {
		testCases := []PushEmailTestCase{
			{
				CaseName:	"Success (New Sub)",
				Sub:		models.Sub{},
				Path:		service.templatesPath,
			},
			{
				CaseName:	"Invalid path (New Sub)",
				Sub:		models.Sub{},
				Path:		"",
			},
			{
				CaseName:	"Success",
				Sub:		models.Sub{
					UUID: TestUid,
				},
				Path:		service.templatesPath,
			},
			{
				CaseName:	"Send Error",
				Sub:		models.Sub{
					UUID: TestUid,
				},
				Path:		service.templatesPath,
				Err:		errors.New("test error"),
			},
			{
				CaseName:	"Send Error (New Sub)",
				Sub:		models.Sub{},
				Path:		service.templatesPath,
				Err:		errors.New("test error"),
			},
		}

		for _, c := range testCases {
			t.Logf("%s: Begin", c.CaseName)

			ctx := context.TODO()

			client.On("Send", message).Return(c.Err)

			service.templatesPath = c.Path

			err := service.PushEmail(ctx, c.Sub)

			a.NoError(err, err)

			client.AssertCalled(t, "Send", message)

			t.Logf("%s: Done", c.CaseName)

			client.ExpectedCalls = []*mock.Call{}
		}
	})

	t.Run("searchSub", func(t *testing.T) {
		testCases := []SearchSubTestCase{
			{
				CaseName:	"Success",
				Sub:		models.Sub{
					UUID: TestUid,
				},
			},
			{
				CaseName:	"Empty data",
				Sub:		models.Sub{},
			},
			{
				CaseName:	"Unknown number",
				Sub:		models.Sub{
					UUID: uuid.New(),
				},
			},
		}

		for _, c := range testCases {
			t.Logf("%s: Begin", c.CaseName)

			ctx := context.TODO()

			s, err := service.searchSub(ctx, c.Sub)

			a.NoError(err, err)
			a.Equal(s.Email, c.Sub.Email)
			a.Equal(s.UUID, c.Sub.UUID)

			t.Logf("%s: Done", c.CaseName)
		}
	})

	t.Run("sendEmail", func(t *testing.T) {
		testCases := []SendEmailTestCase{
			{
				CaseName:	"Success",
				Sub:		models.Sub{},
			},
			{
				CaseName:	"Error",
				Sub:		models.Sub{},
				Err:		errors.New("test error"),
			},
		}

		for _, c := range testCases {
			t.Logf("%s: Begin", c.CaseName)

			ctx := context.TODO()

			client.On("Send", message).Return(c.Err)

			err := service.sendEmail(ctx, c.Sub, message)

			a.NoError(err, err)

			client.AssertCalled(t, "Send", message)

			t.Logf("%s: Done", c.CaseName)

			client.ExpectedCalls = []*mock.Call{}
		}
	})

	t.Run("BuildTemplate", func(t *testing.T) {
		testCases := []BuildTemplateTestCase{
			{
				CaseName:	"Success",
				Sub:		models.Sub{
					UUID: TestUid,
				},
				Path:		service.templatesPath,
			},
			{
				CaseName:	"Success (New Sub)",
				Sub:		models.Sub{},
				Path:		service.templatesPath,
			},
			{
				CaseName:	"Error parse",
				Sub:		models.Sub{},
				Path:		"",
			},
			{
				CaseName:	"Err exec",
				Sub:		models.Sub{
					UUID: TestUid,
				},
				Path:		"../templates/test.html",
			},
			{
				CaseName:	"Err exec (New Sub)",
				Sub:		models.Sub{},
				Path:		"../templates/test.html",
			},
		}

		for _, c := range testCases {
			t.Logf("%s: Begin", c.CaseName)

			service.templatesPath = c.Path

			s, err := service.BuildTemplate(c.Sub)

			a.NoError(err, err)
			a.Equal(s.Email, c.Sub.Email)
			a.Equal(s.UUID, c.Sub.UUID)

			t.Logf("%s: Done", c.CaseName)
		}
	})
}
