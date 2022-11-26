package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"mailganer_test_task/internal/models"
	"mailganer_test_task/internal/transport"
	log "mailganer_test_task/pkg/logger"
	"sync"
	"time"

	"github.com/google/uuid"
	cron "github.com/robfig/cron/v3"
)

// MailingServiceConfig Конфигурация для MailingService
type MailingServiceConfig struct {
	Subs			map[models.Sub]*bytes.Buffer
	EmailClient		*email.Client
	TemplatesPath	string
	ImagePath		string
	Logger			*log.Log
	Message			*email.Message
}

// MailingService Сервис почтовой рассылки
type MailingService struct {
	mux				sync.RWMutex
	sched			*cron.Cron
	subs			map[models.Sub]*bytes.Buffer
	emailClient		*email.Client
	templatesPath	string
	imagePath		string
	logger			*log.Log
	message			*email.Message
}

// NewMailingService Конструктор для MailingService
func NewMailingService(c *MailingServiceConfig) *MailingService {
	return &MailingService{
		message: 		c.Message,
		logger: 		c.Logger,
		imagePath:		c.ImagePath,
		templatesPath:	c.TemplatesPath,
		emailClient:	c.EmailClient,
		subs:			c.Subs,
		mux: 			sync.RWMutex{},
		sched:			cron.New(cron.WithLocation(time.UTC), cron.WithLogger(cron.DefaultLogger)),
	}
}

// WriteOpening Записывает в лог прочтение письма получателем
func (s *MailingService) WriteOpening(ctx context.Context, uid string) {
	ctx = log.ContextWithSpan(ctx, "WriteOpening")
	l := s.logger.WithContext(ctx)

	l.Debug("WriteOpening() started")
	defer l.Debug("WriteOpening() done")

	var fn, ln string
	for sub := range s.subs {
		if sub.UUID.String() == uid {
			fn = sub.Firstname
			ln = sub.Lastname
		}
	}

	//
	l.Infof("User %s %s has read the message", fn, ln)
	return
}

// GetImage Отдает изображения, используемое как индикатор открытия письма//////////////////////////////////////
func (s *MailingService) GetImage(ctx context.Context) ([]byte, error) {
	ctx = log.ContextWithSpan(ctx, "GetImage")
	l := s.logger.WithContext(ctx)

	l.Debug("GetImage() started")
	defer l.Debug("GetImage() done")
	
	result := []byte{}
	
	return result, nil
}

// AddSubTemplate Добавляет подписчика и шаблон письма в кэш
func (s *MailingService) AddSubTemplate(ctx context.Context, sub models.Sub) error {
	ctx = log.ContextWithSpan(ctx, "AddSubTemplate")
	l := s.logger.WithContext(ctx)

	l.Debug("AddSubTemplate() started")
	defer l.Debug("AddSubTemplate() done")

	//
	msgForm, err := s.buildTemplate(ctx, sub, s.templatesPath)
	if err != nil {
		l.Errorf("Unable to build template: %v", err)
		return err
	}

	//
	Id, err := s.sched.AddFunc(formatSched(sub.BirthDay), func(){
		err := s.sendEmail(ctx, sub, msgForm, s.message)
		if err != nil {
			l.Errorf("Unable to send mail in scheduler: %v", err)
		}
	})
	if err != nil {
		l.Errorf("Unable to add func to scheduler: %v", err)
		return err
	}

	//
	sub.CronID = Id
	sub.UUID = uuid.New()
	s.mux.Lock()
	s.subs[sub] = msgForm
	s.mux.Unlock() 

	return nil
}

// RemoveSub Удаляет подписчика из кэша
func (s *MailingService) RemoveSub(ctx context.Context, sub models.Sub) {
	ctx = log.ContextWithSpan(ctx, "RemoveSub")
	l := s.logger.WithContext(ctx)

	l.Debug("RemoveSub() started")
	defer l.Debug("RemoveSub() done")

	//
	s.mux.Lock()
	defer s.mux.Unlock()

	for k := range s.subs {
		if k.BirthDay == sub.BirthDay && k.Firstname == sub.Firstname && k.Lastname == sub.Lastname {
			s.sched.Remove(k.CronID)
			delete(s.subs, k)
			return
		}
	}

	return
}

// PushEmail Отправляет письмо по требованию
func (s *MailingService) PushEmail(ctx context.Context, sub models.Sub) error {
	ctx = log.ContextWithSpan(ctx, "PushEmail")
	l := s.logger.WithContext(ctx)

	l.Debug("PushEmail() started")
	defer l.Debug("PushEmail() done")

	//
	msgForm, ok := s.subs[sub]

	//
	if !ok {
		//
		msgForm, err := s.buildTemplate(ctx, sub, s.templatesPath)
		if err != nil {
			l.Errorf("Unable to build template: %v", err)
			return err
		}

		//
		err = s.sendEmail(ctx, sub, msgForm, s.message)
		if err != nil {
			l.Errorf("Unable to send message: %v", err)
			return err
		}

		return nil
	}

	//
	err := s.sendEmail(ctx, sub, msgForm, s.message)
	if err != nil {
		l.Errorf("Unable to send message: %v", err)
		return err
	}

	return nil
}

// formatSched Рассчитывает время и переводит его в формат планировщика
func formatSched(date time.Time) string {
	// Парсим составляющие даты
	mi := date.Minute()
	h := date.Hour()
	d := date.Day()
	mo := int(date.Month())

	// Вставляем значения в шаблон планировщика
	result := fmt.Sprintf("%v %v %v %v *", mi, h, d, mo)
	return result
}

// buildTemplate Собирает шаблон письма для подписчика и записывает его в буфер
func (s *MailingService) buildTemplate(ctx context.Context, sub models.Sub, templatesPath string) (*bytes.Buffer, error) {
	ctx = log.ContextWithSpan(ctx, "buildTemplate")
	l := s.logger.WithContext(ctx)

	l.Debug("buildTemplate() started")
	defer l.Debug("buildTemplate() done")

	//
	tmplt, err := template.ParseFiles(templatesPath)
	if err != nil {
		l.Errorf("Unable to parse template file: %v", err)
		return	nil, err
	}

	//
	buf := new(bytes.Buffer)

	//
	if err := tmplt.Execute(buf, sub); err != nil {
		l.Errorf("Unable to write template to buffer: %v", err)
		return nil, err
	}

	return buf, nil
}

// sendEmail Отправляет письмо
func (s *MailingService) sendEmail(ctx context.Context, sub models.Sub, msgForm *bytes.Buffer, message *email.Message) error {
	ctx = log.ContextWithSpan(ctx, "sendEmail")
	l := s.logger.WithContext(ctx)

	l.Debug("sendEmail() started")
	defer l.Debug("sendEmail() done")

	//
	msg := message

	//
	msg.ToEmails = append(msg.ToEmails, sub.Email)
	msg.Message = msgForm.String()

	//
	if err := s.emailClient.Send(msg); err != nil {
		l.Errorf("Unable to send message in email client: %v", err)
		return err
	}

	return nil
}
