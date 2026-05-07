package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

type Middlewares struct {
	LanguageFromQuery gin.HandlerFunc
	LanguageFromBody  gin.HandlerFunc
	Auth              gin.HandlerFunc
}

func SetupMiddlewares(auth authClient, logger *slog.Logger) *Middlewares {
	return &Middlewares{
		LanguageFromQuery: LanguageFromQueryMiddleware(logger),
		LanguageFromBody:  LanguageFromBodyMiddleware(logger),
		Auth:              AuthMiddleware(auth, logger),
	}
}
