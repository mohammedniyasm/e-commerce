package bootstrap

import (
	"ecommerce/internal/delivery/http/handler"
	"ecommerce/internal/delivery/http/router"
	"ecommerce/internal/repository/postgres"
	"ecommerce/internal/usecase/auth"
	"log/slog"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Application struct {
	Router *gin.Engine
}

func NewApplication(db *gorm.DB, log *slog.Logger) *Application {
	userRepository := postgres.NewUserRepository(db)
	authUseCase := auth.NewAuthUseCase(userRepository, log)
	authHandler := handler.NewAuthHandler(authUseCase)
	r := router.SetupRouter(log, authHandler)
	return &Application{
		Router: r,
	}
}
