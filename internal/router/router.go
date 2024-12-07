package router

import (
	"github.com/go-chi/chi/v5"
	cMiddleware "github.com/go-chi/chi/v5/middleware"
	"passkeeper/internal/services/user"

	//_ "gofemart/api"
	"passkeeper/internal/config"
	"passkeeper/internal/database"
	"passkeeper/internal/logger"
	"passkeeper/internal/middlewares"
)

// NewRouter конфигурация роутинга приложение
func NewRouter(dbPool *database.DBPool, cnf *config.CliConfig) chi.Router {
	lHandlers := user.NewHandlers(dbPool.DBx, cnf.JWTKeys, cnf.TokenExpiration, cnf.TokenExpiration, cnf.HashKey)
	//authenticator := token.NewAuthenticator(dbPool.DBx, cnf.JWTKeys, cnf.TokenExpiration)
	router := chi.NewRouter()
	// Устанавливаем мидлваре
	router.Use(
		middlewares.JSONHeaders,
		cMiddleware.StripSlashes, // Убираем лишние слеши
		logger.LogRequests,       // Логируем данные запроса
	)
	// Адрес свагера
	/*router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), //The url pointing to API definition
	))*/
	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", lHandlers.RegistrationHandler)
		r.Post("/login", lHandlers.LoginHandler)
	})

	return router
}
