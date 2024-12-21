package router

import (
	"github.com/go-chi/chi/v5"
	cMiddleware "github.com/go-chi/chi/v5/middleware"
	"passkeeper/internal/services/content"
	"passkeeper/internal/services/user"
	"passkeeper/internal/token"

	//_ "passkeeper/api"
	"passkeeper/internal/config"
	"passkeeper/internal/database"
	"passkeeper/internal/logger"
	"passkeeper/internal/middlewares"
)

// NewRouter конфигурация роутинга приложение
func NewRouter(dbPool *database.DBPool, cnf *config.CliConfig) chi.Router {
	lHandlers := user.NewHandlers(dbPool.DBx, cnf.JWTKeys, cnf.TokenExpiration, cnf.TokenExpiration, cnf.HashKey)
	pHandlers := content.NewPasswordService(dbPool.DBx)
	tHandlers := content.NewTextService(dbPool.DBx)
	cHandlers := content.NewCardService(dbPool.DBx)
	fHandlers := content.NewFileService(dbPool.DBx, cnf.UploadPath)
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
		r.Group(registerTextRoutes(tHandlers, authenticator))
		r.Group(registerCardRoutes(cHandlers, authenticator))
		r.Group(registerFileRoutes(fHandlers, authenticator))
	})

	return router
}

// registerPasswordRoutes маршруты с паролями
func registerPasswordRoutes(pHandlers *content.PasswordService, authenticator *token.Authenticator) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(authenticator.Middleware)
		r.Post("/password", pHandlers.SavePasswordHandler)
		r.Put("/password", pHandlers.UpdatePasswordHandler)
		r.Get("/password", pHandlers.GetUserPasswords)
		r.Delete("/password/{id}", pHandlers.DeleteUserPasswords)
	}
}

// registerTextRoutes маршруты с текстами
func registerTextRoutes(pHandlers *content.TextService, authenticator *token.Authenticator) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(authenticator.Middleware)
		r.Post("/text", pHandlers.SaveTextHandler)
		r.Put("/text", pHandlers.UpdateTextHandler)
		r.Get("/text", pHandlers.GetUserTexts)
		r.Delete("/text/{id}", pHandlers.DeleteUserText)
	}
}

// registerCardRoutes маршруты с картами
func registerCardRoutes(pHandlers *content.CardService, authenticator *token.Authenticator) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(authenticator.Middleware)
		r.Post("/card", pHandlers.SaveCardHandler)
		r.Put("/card", pHandlers.UpdateCardHandler)
		r.Get("/card", pHandlers.GetUserCards)
		r.Delete("/card/{id}", pHandlers.DeleteUserCard)
	}
}

// registerFileRoutes маршруты с файлами
func registerFileRoutes(pHandlers *content.FileService, authenticator *token.Authenticator) func(r chi.Router) {
	return func(r chi.Router) {
		r.Use(authenticator.Middleware)
		r.Post("/file", pHandlers.SaveFileHandler)
		r.Put("/file", pHandlers.UpdateFileHandler)
		r.Get("/file", pHandlers.GetUserFiles)
		r.Delete("/file/{id}", pHandlers.DeleteUserFile)
		r.Get("/file/download/{id}", pHandlers.DownloadFileHandler)
	}
}
