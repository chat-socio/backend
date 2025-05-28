package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/chat-socio/backend/configuration"
	"github.com/chat-socio/backend/infrastructure/http"
	"github.com/chat-socio/backend/infrastructure/nats"
	"github.com/chat-socio/backend/infrastructure/postgresql"
	"github.com/chat-socio/backend/infrastructure/redis"
	"github.com/chat-socio/backend/internal/domain"
	"github.com/chat-socio/backend/internal/handler"
	"github.com/chat-socio/backend/internal/middleware"
	"github.com/chat-socio/backend/internal/usecase"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/hertz-contrib/cors"
	"github.com/hertz-contrib/websocket"
	natsjs "github.com/nats-io/nats.go"
)

type Handler struct {
	UserHandler         *handler.UserHandler
	ConversationHandler *handler.ConversationHandler
	Middleware          *middleware.Middleware
	WebSocketHandler    *handler.WebSocketHandler
}

func CreateStream(js natsjs.JetStreamContext) error {
	js.DeleteStream(domain.STREAM_NAME_CONVERSATION)
	js.DeleteStream(domain.STREAM_NAME_WS_MESSAGE)
	_, err := js.AddStream(&natsjs.StreamConfig{
		Name:     domain.STREAM_NAME_CONVERSATION,
		Subjects: []string{domain.SUBJECT_WILDCARD_CONVERSATION},
	})

	if err != nil {
		return err
	}

	_, err = js.AddStream(&natsjs.StreamConfig{
		Name:     domain.STREAM_NAME_WS_MESSAGE,
		Subjects: []string{domain.SUBJECT_WILDCARD_MESSAGE},
	})
	if err != nil {
		return err
	}

	return nil

}

func RunApp() {
	ctx, cancel := context.WithCancel(context.Background())
	// Initialize the database connection
	db, err := postgresql.Connect(ctx, configuration.ConfigInstance.Postgres)
	if err != nil {
		panic(err)
	}

	redisClient := redis.Connect(configuration.ConfigInstance.Redis)

	natsClient := nats.Connect(configuration.ConfigInstance.Nats.Address)
	js, err := natsClient.JetStream()
	if err != nil {
		panic(err)
	}
	//Init websocket
	domain.InitWebSocket()

	// Create stream
	err = CreateStream(js)
	if err != nil {
		panic(err)
	}

	// Initialize repositories
	accountRepository := postgresql.NewAccountRepository(db)
	userRepository := postgresql.NewUserRepository(db)
	sessionRepository := postgresql.NewSessionRepository(db)
	sessionCacheRepository := redis.NewSessionCacheRepository(redisClient)
	userCacheRepository := redis.NewUserCacheRepository(redisClient)
	conversationRepository := postgresql.NewConversationRepository(db)
	messageRepository := postgresql.NewMessageRepository(db)
	userOnlineRepository := postgresql.NewUserOnlineRepository(db)

	// Initialize publisher
	messagePublisher := nats.NewPublisher(js)

	// Initialize use cases
	userUseCase := usecase.NewUserUseCase(accountRepository, userRepository, sessionRepository, sessionCacheRepository, userCacheRepository)
	conversationUseCase := usecase.NewConversationUseCase(conversationRepository, messageRepository, messagePublisher, userOnlineRepository, userRepository)
	userOnlineUseCase := usecase.NewUserOnlineUsecase(userOnlineRepository)

	// Initialize the handler
	handler := &Handler{
		UserHandler: &handler.UserHandler{
			UserUseCase: userUseCase,
		},

		Middleware: middleware.NewMiddleware(sessionCacheRepository, sessionRepository),
		WebSocketHandler: handler.NewWebSocketHandler(&websocket.HertzUpgrader{
			CheckOrigin: func(c *app.RequestContext) bool {
				return true
			},
		}, userOnlineUseCase, userUseCase),
		ConversationHandler: &handler.ConversationHandler{
			ConversationUseCase: conversationUseCase,
			UserUseCase:         userUseCase,
		},
	}

	// Init subscriber
	WsNewMessageSubscriber := nats.NewSubscriber(js, domain.CONSUMER_NAME_WS_MESSAGE_NEW)
	err = WsNewMessageSubscriber.Subscribe(ctx, domain.SUBJECT_NEW_MESSAGE, nats.WrapHandler(conversationUseCase.HandleNewMessage))
	if err != nil {
		panic(err)
	}

	UpdateLastMessageSubscriber := nats.NewQueueSubscriber(js, domain.QUEUE_NAME_WS_MESSAGE_UPDATE_LAST_MESSAGE, domain.CONSUMER_NAME_WS_MESSAGE_UPDATE_LAST_MESSAGE)
	err = UpdateLastMessageSubscriber.Subscribe(ctx, domain.SUBJECT_UPDATE_LAST_MESSAGE_ID, nats.WrapHandler(conversationUseCase.HandleUpdateLastMessageID))
	if err != nil {
		panic(err)
	}

	// Initialize the server
	s := http.NewServer(configuration.ConfigInstance.Server)
	s.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length", "Authorization"},
		AllowCredentials: true,
	}))

	// Set up routes
	SetUpRoutes(s, handler)

	//graceful shutdown
	var signalChan = make(chan os.Signal, 1)
	go func() {
		signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
		<-signalChan
		fmt.Println("Received shutdown signal, shutting down gracefully...")
		WsNewMessageSubscriber.Unsubscribe()
		db.Close()
		redisClient.Close()
		natsClient.Drain()
		cancel()
	}()

	// Start the server
	s.Spin()
}

func SetUpRoutes(s *server.Hertz, handler *Handler) {
	// Route not use auth middleware
	s.POST(("/user/register"), handler.UserHandler.Register)
	s.POST(("/user/login"), handler.UserHandler.Login)
	// Route use auth middleware
	authGroup := s.Group("/auth")
	authGroup.Use(handler.Middleware.AuthMiddleware())
	authGroup.GET("/user/info", handler.UserHandler.GetMyInfo)
	authGroup.GET("/user/search", handler.UserHandler.GetListUser)
	// Conversation
	authGroup.GET("/conversation", handler.ConversationHandler.GetListConversation)
	authGroup.POST("/conversation", handler.ConversationHandler.CreateConversation)
	authGroup.GET("/conversation/:conversation_id", handler.ConversationHandler.GetConversationByID)

	// Message
	authGroup.POST("/message", handler.ConversationHandler.SendMessage)
	authGroup.GET("/message", handler.ConversationHandler.GetListMessage)

	s.GET("/ws", handler.WebSocketHandler.HandleWebsocket)
}
