package services

import (
	"context"
	"fmt"
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
	subs			map[uuid.UUID]models.Sub
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
		mux: 			sync.RWMutex{},
		sched:			cron.New(cron.WithLocation(time.UTC), cron.WithLogger(cron.DefaultLogger)),
		subs:			make(map[uuid.UUID]models.Sub),
	}
}

// WriteOpening Записывает в лог прочтение письма получателем
func (s *MailingService) WriteOpening(ctx context.Context, uid uuid.UUID) {
	ctx = log.ContextWithSpan(ctx, "WriteOpening")
	l := s.logger.WithContext(ctx)

	l.Debug("WriteOpening() started")
	defer l.Debug("WriteOpening() done")

	// Находим подписчика в кэше
	sub := s.subs[uid]

	// Сообщение о прочтении адресатом
	l.Infof("User %s %s has read the message", sub.Firstname, sub.Lastname)
}

// AddSubTemplate Добавляет подписчика и шаблон письма в кэш и планировщик
func (s *MailingService) AddSubTemplate(ctx context.Context, sub models.Sub) error {
	ctx = log.ContextWithSpan(ctx, "AddSubTemplate")
	l := s.logger.WithContext(ctx)

	l.Debug("AddSubTemplate() started")
	defer l.Debug("AddSubTemplate() done")

	// Проверяем, есть ли подписчик в кэше
	_, err := s.searchSub(ctx, uuid.Nil, sub)
	if err == nil {
		// Если есть, возвращаем ошибку
		return l.RErrorf("sub has already in mailing base.")
	}

	// Генерируем уникальный номер подписчику
	sub.UUID = uuid.New()
	sub.Uid = sub.UUID.String()

	// Создаем шаблон сообщения
	err = sub.BuildTemplate(ctx, s.templatesPath)
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

	l.Info("Sub has been saved")

	return nil
}

// RemoveSub Ищет подписчика в кэше и удаляет его
func (s *MailingService) RemoveSub(ctx context.Context, uid uuid.UUID, sub models.Sub) error {
	ctx = log.ContextWithSpan(ctx, "RemoveSub")
	l := s.logger.WithContext(ctx)

	l.Debug("RemoveSub() started")
	defer l.Debug("RemoveSub() done")

	// Ищем подписчика
	result, err := s.searchSub(ctx, uid, sub)
	if err != nil {
		return err
	}

	// Удаляем подписчика из системы планирования и из кэша
	s.sched.Remove(result.CronID)
	s.mux.Lock()
	delete(s.subs, result.UUID)
	s.mux.Unlock()

	l.Infof("Sub number %s has been deleted", uid.String())

	return nil
}

// PushEmail Отправляет письмо по требованию
func (s *MailingService) PushEmail(ctx context.Context, sub models.Sub) error {
	ctx = log.ContextWithSpan(ctx, "PushEmail")
	l := s.logger.WithContext(ctx)

	l.Debug("PushEmail() started")
	defer l.Debug("PushEmail() done")

	// Ищем подписчика
	result, err := s.searchSub(ctx, uuid.Nil, sub)
	if err != nil {
		// Если не нашли - создаем ему шаблон
		err := sub.BuildTemplate(ctx, s.templatesPath)
		if err != nil {
			l.Errorf("Unable to build template: %v", err)
			return err
		}
	}

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
func (s *MailingService) searchSub(ctx context.Context, uid uuid.UUID, sub models.Sub) (models.Sub, error) {
	ctx = log.ContextWithSpan(ctx, "sendEmail")
	l := s.logger.WithContext(ctx)

	l.Debug("sendEmail() started")
	defer l.Debug("sendEmail() done")

	// Если входные данные пусты - сразу возвращаем ошибку и записываем в лог
	if (uid == uuid.Nil) && (sub == models.Sub{}) {
		return models.Sub{}, l.RError("invalid incoming data")
	}

	// Если уникальный номер есть, можем быстро найти по нему
	if (uid != uuid.Nil) {
		// Доступ к кэшу по ключу
		s.mux.RLock()
		result, ok := s.subs[uid]
		s.mux.RUnlock()

		// Если есть - возвращаем, если нет - отдаем ошибку и записываем в лог
		if !ok {
			return models.Sub{}, l.RErrorf("sub number %s not found", uid.String())
		} else {
			return result, nil
		}

	}

	// Итерируемся по кэшу
	s.mux.RLock()
	defer s.mux.RUnlock()
	for _, v := range s.subs {
		// Если поля совпадают - возвращаем значение кэша
		if v.BirthDay == sub.BirthDay && v.Firstname == sub.Firstname && v.Lastname == sub.Lastname && v.Email == sub.Email {
			return v, nil
		}
	}

	//Если не нашли - отдаем ошибку и записываем в лог
	return models.Sub{}, l.RError("sub not found")
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
