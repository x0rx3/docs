package transport

import "github.com/gin-gonic/gin"

type AuthHandler interface {
	Login(*gin.Context)
	Logout(*gin.Context)
}

type RegistrationHandler interface{ Register(*gin.Context) }

type DocumentHandler interface {
	AddDocument(*gin.Context)
	GetDocument(ctx *gin.Context)
	ListDocuments(ctx *gin.Context)
	DeleteDocument(ctx *gin.Context)
}
