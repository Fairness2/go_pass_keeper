package router

import (
	"github.com/go-chi/chi/v5"
	cMiddleware "github.com/go-chi/chi/v5/middleware"
	"passkeeper/internal/services/content"
	"passkeeper/internal/services/user"
	"passkeeper/internal/token"

	//_ "gofemart/api"
	"passkeeper/internal/config"
	"passkeeper/internal/database"
	"passkeeper/internal/logger"
	"passkeeper/internal/middlewares"
)

// NewRouter конфигурация роутинга приложение
func NewRouter(dbPool *database.DBPool, cnf *config.CliConfig) chi.Router {
	lHandlers := user.NewHandlers(dbPool.DBx, cnf.JWTKeys, cnf.TokenExpiration, cnf.TokenExpiration, cnf.HashKey)
	pHandlers := content.NewPasswordService(dbPool.DBx, cnf.EncryptKeys)
	authenticator := token.NewAuthenticator(dbPool.DBx, cnf.JWTKeys, cnf.TokenExpiration)

	router := chi.NewRouter()
	// Устанавливаем мидлваре
	router.Use(
		middlewares.JSONHeaders,
		cMiddleware.StripSlashes, // Убираем лишние слеши
		cMiddleware.Compress(5, "gzip", "deflate"),
		logger.LogRequests, // Логируем данные запроса
	)
	// Адрес свагера
	/*router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), //The url pointing to API definition
	))*/
	router.Route("/api/user", func(r chi.Router) {
		r.Post("/register", lHandlers.RegistrationHandler)
		r.Post("/login", lHandlers.LoginHandler)
	})
	router.Route("/api/content", func(r chi.Router) {
		r.Group(registerPasswordRoutes(pHandlers, authenticator))
	})

	return router
}

// registerRoutesWithAuth маршруты с аутентификацией
func registerPasswordRoutes(pHandlers *content.PasswordService, authenticator *token.Authenticator) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(authenticator.Middleware)
		r.Post("/password", pHandlers.SavePasswordHandler)
		r.Put("/password", pHandlers.UpdatePasswordHandler)
		r.Get("/password", pHandlers.GetUserPasswords)
	}
}
