package services

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"sync"
	"time"
	"mailganer_test_task/internal/models"
	"mailganer_test_task/internal/transport"
	log "mailganer_test_task/pkg/logger"

	"github.com/google/uuid"
	cron "github.com/robfig/cron/v3"
)

// Client Интерфейс к email клиенту
type Client interface {
	Send(msg *email.Message) error
}

// Cron Интерфейс к планировщику
type Cron interface {
	AddFunc(spec string, cmd func()) (cron.EntryID, error)
	AddJob(spec string, cmd cron.Job) (cron.EntryID, error)
	Entries() []cron.Entry
	Entry(id cron.EntryID) cron.Entry
	Location() *time.Location
	Remove(id cron.EntryID)
	Run()
	Schedule(schedule cron.Schedule, cmd cron.Job) cron.EntryID
	Start()
	Stop() context.Context
}

// MailingServiceConfig Конфигурация для MailingService
type MailingServiceConfig struct {
	EmailClient		Client
	TemplatesPath	string			`env:"APP_TEMPLATES_PATH"`
	Host			string			`env:"APP_HOST"`
	Port			string			`env:"APP_PORT"`
	Logger			*log.Log
	Message			*email.Message
}

// MailingService Сервис почтовой рассылки
type MailingService struct {
	mux				sync.RWMutex
	sched			Cron
	subs			map[uuid.UUID]models.Sub
	emailClient		Client
	templatesPath	string
	host			string
	port			string
	logger			*log.Log
	message			*email.Message
}

// NewMailingService Конструктор для MailingService
func NewMailingService(c *MailingServiceConfig) *MailingService {
	return &MailingService{
		message: 		c.Message,
		logger: 		c.Logger,
		port: 			c.Port,
		host: 			c.Host,
		templatesPath:	c.TemplatesPath,
		emailClient:	c.EmailClient,
		mux: 			sync.RWMutex{},
		sched:			cron.New(cron.WithLocation(time.UTC), cron.WithLogger(cron.DefaultLogger)),
		subs:			make(map[uuid.UUID]models.Sub),
	}
}

// WriteOpening Записывает в лог прочтение письма получателем
func (s *MailingService) WriteOpening(ctx context.Context, uid uuid.UUID) error {
	ctx = log.ContextWithSpan(ctx, "WriteOpening")
	l := s.logger.WithContext(ctx)

	l.Debug("WriteOpening() started")
	defer l.Debug("WriteOpening() done")

	// Находим подписчика в кэше
	s.mux.RLock()
	sub, ok := s.subs[uid]
	s.mux.RUnlock()

	// Сообщение о прочтении адресатом
	if ok {
		l.Infof("User %s %s has read the message", sub.Firstname, sub.Lastname)

		return nil
	}

	return l.RError("data not found")
}

// AddSubTemplate Добавляет подписчика и шаблон письма в кэш и планировщик
func (s *MailingService) AddSubTemplate(ctx context.Context, sub models.Sub) error {
	ctx = log.ContextWithSpan(ctx, "AddSubTemplate")
	l := s.logger.WithContext(ctx)

	l.Debug("AddSubTemplate() started")
	defer l.Debug("AddSubTemplate() done")

	// Проверяем, есть ли подписчик в кэше
	_, err := s.searchSub(ctx, sub)
	if err == nil {
		// Если есть, возвращаем ошибку
		return l.RError("sub has already in mailing base.")
	}

	// Генерируем уникальный номер подписчику
	sub.UUID = uuid.New()

	// Создаем шаблон сообщения
	sub, err = s.BuildTemplate(sub)
	if err != nil {
		l.Errorf("Unable to build template: %v", err)
		return err
	}

	// Добавляем функцию рассылки по дням рождения в планировщик
	Id, err := s.sched.AddFunc(formatSched(sub.BirthDay), func(){
		err := s.sendEmail(ctx, sub, s.message)
		if err != nil {
			l.Errorf("Unable to send mail in scheduler: %v", err)
		}
	})
	if err != nil {
		l.Errorf("Unable to add func to scheduler: %v", err)
		return err
	}

	// Записываем Id задачи планировщика на случай удаления
	sub.CronID = Id
	
	// Записываем подписчика в кэш
	s.mux.Lock()
	s.subs[sub.UUID] = sub
	s.mux.Unlock() 

	l.Infof("Sub has been saved with UUID: %s", sub.UUID)

	return nil
}

// RemoveSub Ищет подписчика в кэше и удаляет его
func (s *MailingService) RemoveSub(ctx context.Context, sub models.Sub) error {
	ctx = log.ContextWithSpan(ctx, "RemoveSub")
	l := s.logger.WithContext(ctx)

	l.Debug("RemoveSub() started")
	defer l.Debug("RemoveSub() done")

	// Ищем подписчика
	result, err := s.searchSub(ctx, sub)
	if err != nil {
		return err
	}

	// Удаляем подписчика из системы планирования и из кэша
	s.sched.Remove(result.CronID)
	s.mux.Lock()
	delete(s.subs, result.UUID)
	s.mux.Unlock()

	l.Infof("Sub number %s has been deleted", sub.UUID.String())

	return nil
}

// PushEmail Отправляет письмо по требованию
func (s *MailingService) PushEmail(ctx context.Context, sub models.Sub) error {
	ctx = log.ContextWithSpan(ctx, "PushEmail")
	l := s.logger.WithContext(ctx)

	l.Debug("PushEmail() started")
	defer l.Debug("PushEmail() done")

	// Ищем подписчика
	result, err := s.searchSub(ctx, sub)
	if err != nil {
		// Если не нашли - создаем ему шаблон
		result, err = s.BuildTemplate(sub)
		if err != nil {
			l.Errorf("Unable to build template: %v", err)
			return err
		}
	}

	// Запишем пользователя в кэш, но без запланированной рассылки
	s.mux.Lock()
	s.subs[result.UUID] = sub
	s.mux.Unlock()

	// Отправляем сообщение
	err = s.sendEmail(ctx, result, s.message)
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

// searchSub Ищет подписчика в кэше
func (s *MailingService) searchSub(ctx context.Context, sub models.Sub) (models.Sub, error) {
	ctx = log.ContextWithSpan(ctx, "searchSub")
	l := s.logger.WithContext(ctx)

	l.Debug("searchSub() started")
	defer l.Debug("searchSub() done")

	// Ищем по уникальному номеру
	uid := sub.UUID
	switch uid {

	// Если входные данные пусты - сразу возвращаем ошибку и записываем в лог
	case uuid.Nil:
		return sub, l.RError("sub not found")

	// Получаем значение кэша по ключу
	default:
		s.mux.RLock()
		result, ok := s.subs[uid]
		s.mux.RUnlock()

		// Если есть - возвращаем, если нет - отдаем ошибку и записываем в лог
		if !ok {
			return sub, l.RError("sub not found")
		} else {
			return result, nil
		}
	}
}

// sendEmail Отправляет письмо
func (s *MailingService) sendEmail(ctx context.Context, sub models.Sub, message *email.Message) error {
	ctx = log.ContextWithSpan(ctx, "sendEmail")
	l := s.logger.WithContext(ctx)

	l.Debug("sendEmail() started")
	defer l.Debug("sendEmail() done")

	// Берем шаблон сообщения
	msg := message

	// Добавляем к нему кастомные параметры
	msg.ToEmails = append(msg.ToEmails, sub.Email)
	msg.Message = sub.Template.String()

	// Отправляем сообщение
	if err := s.emailClient.Send(msg); err != nil {
		l.Errorf("Unable to send message in email client: %v", err)
		return err
	}

	l.Info("Message sent")

	return nil
}

// BuildTemplate Собирает шаблон письма для подписчика и записывает его в буфер
func (s *MailingService) BuildTemplate(sub models.Sub) (models.Sub, error) {
	// Парсим html файл в шаблон
	tmplt, err := template.ParseFiles(s.templatesPath)
	if err != nil {
		return sub, err
	}

	// Создаем байтовый буфер
	buf := new(bytes.Buffer)

	// Если пользователь не записан в рассылки, нужно создать ему uuid для ссылки на изображение
	if sub.UUID == uuid.Nil {
		sub.UUID = uuid.New()
	}

	// Маппим данные для шаблона
	data := models.TemplateData{
		Firstname:	sub.Firstname,
		Lastname:	sub.Lastname,
		URL:		s.host+":"+s.port+"/"+sub.UUID.String(),
	}

	// Записываем данные подписчика в шаблон, а шаблон в буфер
	if err := tmplt.Execute(buf, data); err != nil {
		return sub, err
	}

	// Буфер с готовым шаблоном закрепляем за подписчиком
	sub.Template = buf

	return sub, nil
}
