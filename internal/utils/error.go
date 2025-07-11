package utils

import (
	"docs/internal/transport/http/dto"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrorInvalidAuthData   = errors.New("invalid data")
	ErrorAuthFailed        = errors.New("authorization failed")
	ErrorInvalidToken      = errors.New("invalid token")
	ErrorEmptyUUID         = errors.New("uuid cant be empty")
	ErrorNotFound          = errors.New("not found")
	ErrorInvalidAdminToken = errors.New("invalid admin token")
	ErrorInvalidPassword   = errors.New("invalid password")
	ErrorInvalidLogin      = errors.New("invalid login")
	ErrorInvalidGrant      = errors.New("invalid grant")
	ErrorFilterFormat      = errors.New("invalid filter format")
	ErrorLimitFormat       = errors.New("invalid limit format")
	ErrorEmptyFile         = errors.New("empty file")
	ErrorCacheValue        = errors.New("unxpected type from cache")
	ErrorLoginAlradyExists = errors.New("user with such a login already has")
	ErrorNoAccess          = errors.New("access denied")
)

var errorStatusMap = map[error]int{
	ErrorAuthFailed:        http.StatusUnauthorized,
	ErrorEmptyFile:         http.StatusBadRequest,
	ErrorFilterFormat:      http.StatusBadRequest,
	ErrorEmptyUUID:         http.StatusBadRequest,
	ErrorInvalidGrant:      http.StatusBadRequest,
	ErrorInvalidAuthData:   http.StatusBadRequest,
	ErrorInvalidToken:      http.StatusBadRequest,
	ErrorInvalidAdminToken: http.StatusBadRequest,
	ErrorInvalidLogin:      http.StatusBadRequest,
	ErrorInvalidPassword:   http.StatusBadRequest,
	ErrorNotFound:          http.StatusNotFound,
	ErrorLoginAlradyExists: http.StatusConflict,
	ErrorNoAccess:          http.StatusForbidden,
}

func CaseError(ctx *gin.Context, err error) {
	for target, status := range errorStatusMap {
		if errors.Is(err, target) {
			ctx.JSON(status, dto.ErrorResponse{Error: dto.Error{
				Code: status,
				Text: err.Error(),
			}})
			return
		}
	}

	ctx.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: dto.Error{
			Code: http.StatusInternalServerError,
			Text: "internal server error",
		},
	})
}
