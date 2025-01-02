package router

import (
	"github.com/go-chi/chi/v5"
	cMiddleware "github.com/go-chi/chi/v5/middleware"
	"passkeeper/internal/repositories"
	"passkeeper/internal/services/content"
	"passkeeper/internal/services/user"
	"passkeeper/internal/token"
	//_ "passkeeper/api"
	"passkeeper/internal/config"
	"passkeeper/internal/database"
	"passkeeper/internal/jsonmiddleware"
	"passkeeper/internal/logger"
)

// NewRouter конфигурация роутинга приложение
func NewRouter(dbPool *database.DBPool, cnf *config.CliConfig) chi.Router {
	dbAdapter := repositories.NewDBAdapter(dbPool.DBx)
	userCnf := user.HandlerConfig{
		Repository:             repositories.NewUserRepository(dbAdapter),
		JwtKeys:                cnf.JWTKeys,
		TokenExpiration:        cnf.TokenExpiration,
		RefreshTokenExpiration: cnf.TokenExpiration,
		HashKey:                cnf.HashKey,
	}
	lHandlers := user.NewHandlers(userCnf)
	pHandlers := content.NewPasswordService(repositories.NewPasswordRepository(dbAdapter))
	tHandlers := content.NewTextService(repositories.NewTextRepository(dbAdapter))
	cHandlers := content.NewCardService(repositories.NewCardRepository(dbAdapter))
	fHandlers := content.NewFileService(repositories.NewFileRepository(dbAdapter), cnf.UploadPath)
	authenticator := token.NewAuthenticator(dbAdapter, cnf.JWTKeys, cnf.TokenExpiration)

	router := chi.NewRouter()
	// Устанавливаем мидлваре
	router.Use(
		jsonmiddleware.JSONHeaders,
		cMiddleware.StripSlashes, // Убираем лишние слеши
		cMiddleware.Compress(5, "gzip", "deflate"),
		logger.LogRequests, // Логируем данные запроса
	)
	// Адрес свагера
	/*router.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"), //The url pointing to API definition
	))*/
	router.Route("/api/user", func(r chi.Router) {
		r.Group(lHandlers.RegisterRoutes())
	})
	router.Route("/api/content", func(r chi.Router) {
		r.Group(pHandlers.RegisterRoutes(authenticator.Middleware))
		r.Group(tHandlers.RegisterRoutes(authenticator.Middleware))
		r.Group(cHandlers.RegisterRoutes(authenticator.Middleware))
		r.Group(fHandlers.RegisterRoutes(authenticator.Middleware))
	})

	return router
}
