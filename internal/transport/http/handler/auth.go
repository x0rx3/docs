package handler

import (
	"docs/internal/service"
	"docs/internal/transport/http/dto"
	"docs/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	docsService service.AuthService
}

func NewAuth(docsService service.AuthService) *Auth {
	return &Auth{
		docsService: docsService,
	}
}

// Login godoc
// @Summary      Login
// @Description  Login with login & password
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        data body dto.AuthData true "docs data"
// @Success      200  	{object}  dto.SuccessResponse{response=dto.Token}  "desc"
// @Router       /auth [post]
func (inst *Auth) Login(ctx *gin.Context) {
	docs := &dto.AuthData{}
	if err := ctx.ShouldBindBodyWithJSON(docs); err != nil {
		ctx.JSON(http.StatusBadRequest, dto.SuccessResponse{Response: "bad request"})
		return
	}

	token, err := inst.docsService.Login(ctx, docs.Login, docs.Pswd)
	if err != nil {
		utils.CaseError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Response: dto.Token{Token: token.AccessToken}})

}

// Logout godoc
// @Summary      Logout
// @Description  Logout
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param token path string true "Access Token"
// @Success      200  	{object}  dto.SuccessResponse{response=string}  "desc"
// @Router       /auth/{token} [delete]
func (inst *Auth) Logout(ctx *gin.Context) {
	token := ctx.Param("token")
	if token == "" {
		utils.CaseError(ctx, utils.ErrorInvalidToken)
		return
	}

	if err := inst.docsService.Logout(ctx, token); err != nil {
		utils.CaseError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Response: map[string]bool{
		token: true,
	}})
}
