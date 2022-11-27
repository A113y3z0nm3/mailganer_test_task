package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"mailganer_test_task/internal/config"
	"mailganer_test_task/internal/handlers"
	"mailganer_test_task/internal/services"
	"mailganer_test_task/internal/transport"
	clog "mailganer_test_task/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// Загружаем конфиг
	c, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Загружаем логгер
	logger, err := clog.InitLogger(c.Log)
	if err != nil {
		log.Fatal(err)
	}

	// Создаем контекст с логированием
	ctx := clog.ContextWithTrace(context.Background(), "main")
	ctx = clog.ContextWithSpan(ctx, "main")
	l := logger.WithContext(ctx)

	// Создаем email клиент
	emailClient, err := email.NewClient(c.Email)
	if err != nil {
		log.Fatal(err)
	}
	l.Info("created email client")

	// Инициализируем роутер GIN
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Создаем сервис
	c.Service.Message = c.Message
	c.Service.EmailClient = emailClient
	c.Service.Logger = logger
	mailingService := services.NewMailingService(c.Service)
	l.Info("created mailing service")

	// Регистрируем обработчик
	c.Handler.Router = router
	c.Handler.MailingService = mailingService
	c.Handler.Logger = logger
	handlers.RegisterMailingHandler(c.Handler)
	l.Info("mailing handler has been registered")

	// Создаем сервер
	server := &http.Server{
		Addr: c.Service.Host+":"+c.Service.Port,
		Handler: router,
	}

	// Запускаем сервер
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to initialize server: %v\n", err)
		}
	}()
	l.Infof("server listening on port %v", c.Service.Port)

	// Graceful shutdown
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	l.Info("shutting down server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v\n", err)
	}

	l.Info("shutting down logger")
	l.LogGracefulShutdown()

	log.Println("successfully")
	<-ctx.Done()
}
