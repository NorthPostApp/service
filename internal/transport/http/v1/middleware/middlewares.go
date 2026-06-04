package middleware

import (
	"log/slog"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

type Middlewares struct {
	LanguageFromQuery gin.HandlerFunc
	LanguageFromBody  gin.HandlerFunc
	Auth              gin.HandlerFunc
}

func SetupMiddlewares(
	authType MiddlewareType,
	auth *auth.Client,
	logger *slog.Logger) *Middlewares {
	return &Middlewares{
		LanguageFromQuery: LanguageFromQueryMiddleware(logger),
		LanguageFromBody:  LanguageFromBodyMiddleware(logger),
		Auth:              AuthMiddleware(authType, auth, logger),
	}
}
