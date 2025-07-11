package handler

import (
	"docs/internal/service"
	"docs/internal/transport/http/dto"
	"docs/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Registration struct {
	registerService service.RegistrationService
}

func NewRegistration(registerService service.RegistrationService) *Registration {
	return &Registration{
		registerService: registerService,
	}
}

// Register godoc
// @Summary Registration new user
// @Description Registration new user
// @Tags Registration
// @Accept json
// @Produce json
// @Param data body dto.Registration true "Regestration data"
// @Success 201 {object} dto.SuccessResponse{response=string} "desc"
// @Router /register [post]
func (inst *Registration) Register(ctx *gin.Context) {
	regData := &dto.Registration{}
	if err := ctx.ShouldBindBodyWithJSON(regData); err != nil {
		utils.CaseError(ctx, utils.ErrorInvalidAuthData)
		return
	}

	if err := inst.registerService.Register(ctx, regData.Token, regData.Login, regData.Password); err != nil {
		utils.CaseError(ctx, err)
		return
	}

	ctx.JSON(http.StatusCreated, &dto.SuccessResponse{Response: "test"})
}
