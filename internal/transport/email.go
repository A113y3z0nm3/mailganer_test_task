package email

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
	"time"
)

const (
	delimiter = "**=myohmy689407924327"
	boundary  = "my-boundary-779"
)

// File структура для отправки файлов
type File struct {
	Body []byte
	Name string
}

// Message структура сообщения
type Message struct {
	Files      []File   							// Список файлов
	Message    string   							// Текст сообщения
	Subject    string   	`env:"MESSAGE_SUBJECT"`	// Заголовок
	ToEmails   []string 							// Список получателей
	FromEmail  string   	`env:"MESSAGE_FROM"`	// Отправитель
	CarbonCopy []string 	`env:"MESSAGE_COPY"`	// Список добавленных в копию
}

// Config конфигурация для клиента
type EmailConfig struct {
	Host       string	`env:"EMAIL_HOST"`
	Port       string	`env:"EMAIL_PORT"`
	Username   string	`env:"EMAIL_USERNAME"`
	Password   string	`env:"EMAIL_PASSWORD"`
	Timeout    int		`env:"EMAIL_TIMEOUT"`
	TlsEnabled bool		`env:"EMAIL_TLS_BOOL"`
}

// Client реализует клиент для отправки сообщений
type Client struct {
	host       string
	port       string
	serverName string
	username   string
	password   string
	timeout    int
	tlsEnabled bool
}

// NewClient создает Client
func NewClient(c *EmailConfig) (*Client, error) {
	// Проверяем ключевые поля, чтобы они не были пустыми
	if err := validateConfig(c); err != nil {
		return nil, err
	}

	client := &Client{
		host:       c.Host,
		port:       c.Port,
		serverName: fmt.Sprintf("%s:%s", c.Host, c.Port),
		username:   c.Username,
		password:   c.Password,
		timeout:    c.Timeout,
		tlsEnabled: c.TlsEnabled,
	}

	return client, nil
}

// Создает подключение с включенным tls
func (c *Client) createTlsConn() (net.Conn, error) {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         c.host,
	}

	// Настраиваем timeout
	dialer := &net.Dialer{
		Timeout: time.Duration(c.timeout) * time.Second,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", c.serverName, tlsConfig)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Создание обычного подключения к smtp серверу
func (c *Client) createConn() (net.Conn, error) {
	log.Println(c.serverName)
	conn, err := net.DialTimeout("tcp", c.serverName, time.Duration(c.timeout)*time.Second)

	if err != nil {
		return nil, err
	}

	return conn, nil
}

// Проверяет ключевые поля конфигурации
func validateConfig(c *EmailConfig) error {
	if len(c.Host) == 0 {
		return fmt.Errorf("empty host")
	}

	if len(c.Port) == 0 {
		return fmt.Errorf("empty port")
	}

	if len(c.Username) == 0 {
		return fmt.Errorf("empty username")
	}

	// Если timeout не передан, то выставляем default значение
	if c.Timeout == 0 {
		c.Timeout = 5
	}

	return nil
}

func (c *Client) Send(msg *Message) error {
	var conn net.Conn
	var err error

	conn, err = c.createConn()

	if err != nil {
		return err
	}

	if err = c.create(conn, msg); err != nil {
		return err
	}

	return nil
}

// Создает итоговое письмо
func (c *Client) create(conn net.Conn, msg *Message) error {
	client, err := smtp.NewClient(conn, c.host)
	if err != nil {
		return err
	}

	if c.tlsEnabled {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         c.host,
		}

		if err = client.StartTLS(tlsConfig); err != nil {
			return err
		}
	}

	//auth := smtp.PlainAuth("", c.username, c.password, c.host)

	// Производим авторизацию
	if err = client.Auth(LoginAuth(c.username, c.password)); err != nil {
		return err
	}

	// Добавляем отправителя
	if err = client.Mail(msg.FromEmail); err != nil {
		return err
	}

	log.Println("GO")
	// Добавляем получателей
	for _, to := range msg.ToEmails {
		if err = client.Rcpt(to); err != nil {
			return err
		}
	}

	// Получаем буфер для записи содержимого письма
	writer, err := client.Data()
	if err != nil {
		return err
	}

	sample, err := preprocessData(msg)
	if err != nil {
		return err
	}

	log.Println("GO2")

	// if err = smtp.SendMail(c.host+":"+c.port, auth, msg.FromEmail, msg.ToEmails, sample); err != nil {
	// 	return err
	// }

	// Записываем данные в тело письма
	if _, err = writer.Write(sample); err != nil {
		return err
	}

	if err = writer.Close(); err != nil {
		return err
	}

	// Отправляем сообщение и закрываем соединение
	if err = client.Quit(); err != nil {
		return err
	}

	return nil
}

// Создает письмо
func createSample(msg *Message) string {
	sample := fmt.Sprintf("From: %s\r\n", msg.FromEmail)
	sample += fmt.Sprintf("To: %s\r\n", strings.Join(msg.ToEmails, ";"))

	if len(msg.CarbonCopy) > 0 {
		sample += fmt.Sprintf("Cc: %s\r\n", strings.Join(msg.CarbonCopy, ";"))
	}
	sample += fmt.Sprintf("Subject: %s\r\n", msg.Subject)

	sample += "MIME-Version: 1.0\r\n"
	sample += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimiter)

	sample += fmt.Sprintf("\r\n--%s\r\n", delimiter)
	sample += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	sample += "Content-Transfer-Encoding: 7bit\r\n"
	sample += fmt.Sprintf("\r\n%s\r\n", msg.Message)

	for i := 0; i < len(msg.Files); i++ {
		sample += fmt.Sprintf("\r\n--%s\r\n", delimiter)
		sample += "Content-Type: text/plain; charset=\"utf-8\"\r\n"
		sample += "Content-Transfer-Encoding: base64\r\n"
		sample += "Content-Disposition: attachment;filename=\"" + msg.Files[i].Name + "\"\r\n"
		sample += "\r\n" + base64.StdEncoding.EncodeToString(msg.Files[i].Body)
	}

	return sample
}

// Подготавливает данные для отправки
func preprocessData(msg *Message) ([]byte, error) {

	var buf bytes.Buffer
	var err error

	_, err = buf.WriteString("MIME-Version: 1.0\r\n")
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(fmt.Sprintf("From: %s\r\n", msg.FromEmail))
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.ToEmails, ";")))
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	if err != nil {
		return nil, err
	}

	_, err = buf.WriteString(fmt.Sprintf("\r\n%s", msg.Message))
	if err != nil {
		return nil, err
	}

	// _, err = buf.WriteString(fmt.Sprintf("\r\n--%s", boundary))
	// if err != nil {
	// 	return nil, err
	// }

	// _, err = buf.WriteString("--")
	// if err != nil {
	// 	return nil, err
	// }

	return buf.Bytes(), nil
}
