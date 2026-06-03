package middleware

import (
	"log/slog"

	"cloud.google.com/go/firestore"
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
	db *firestore.Client,
	logger *slog.Logger) *Middlewares {
	return &Middlewares{
		LanguageFromQuery: LanguageFromQueryMiddleware(logger),
		LanguageFromBody:  LanguageFromBodyMiddleware(logger),
		Auth:              AuthMiddleware(authType, auth, db, logger),
	}
}
