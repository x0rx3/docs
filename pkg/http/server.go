package http

import (
	"docs/docs"
	"docs/internal/transport"
	"docs/internal/transport/http/handler"
	"docs/pkg/service"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

type Server struct {
	eng             *gin.Engine
	authHandler     transport.AuthHandler
	registerHandler transport.RegistrationHandler
	documentHandler transport.DocumentHandler
}

func NewServer(log *zap.Logger, serviceCollector *service.ServiceCollector) *Server {
	gin.SetMode(gin.DebugMode)
	return &Server{
		eng:             gin.New(),
		authHandler:     handler.NewAuth(serviceCollector.AuthService),
		registerHandler: handler.NewRegistration(serviceCollector.RegistrationService),
		documentHandler: handler.NewDocuments(log, serviceCollector.DocumentService),
	}
}

func (inst *Server) Start(address, port string) error {
	docs.SwaggerInfo.Title = "API Swagger"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "127.0.0.1" + ":" + port
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Schemes = []string{"http"}

	inst.eng.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	apiGroup := inst.eng.Group("/api")

	// auth routes
	apiGroup.POST("/auth", inst.authHandler.Login)
	apiGroup.DELETE("/auth/:token", inst.authHandler.Logout)

	// register routes
	apiGroup.POST("/register", inst.registerHandler.Register)

	// documents routes
	apiGroup.POST("/docs", inst.documentHandler.AddDocument)
	apiGroup.GET("/docs/:uuid", inst.documentHandler.GetDocument)
	apiGroup.HEAD("/docs/:uuid", inst.documentHandler.GetDocument)
	apiGroup.GET("/docs", inst.documentHandler.ListDocuments)
	apiGroup.HEAD("/docs", inst.documentHandler.ListDocuments)
	apiGroup.DELETE("/docs/:uuid", inst.documentHandler.DeleteDocument)

	return inst.eng.Run(address + ":" + port)
}
