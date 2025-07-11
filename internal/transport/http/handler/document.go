package handler

import (
	"docs/internal/model"
	"docs/internal/service"
	"docs/internal/transport/http/dto"
	"docs/internal/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Document struct {
	log        *zap.Logger
	docService service.DocumentService
}

func NewDocuments(log *zap.Logger, docService service.DocumentService) *Document {
	return &Document{log, docService}
}

// AddDocument godoc
// @Summary Add Document
// @Description Add new document
// @Tags Document
// @Produce json
// @Accept mpfd
// @Param meta formData string true "Document meta data (JSON)" example({"name":"photo.jpg","file":true,"public":false,"token":"sfuqwejqjoiu93e29","mime":"image/jpg","grant":["login1","login2"]})
// @Param json formData string false "Extantion data for document (JSON)" example({"key":"value"})
// @Param file formData file false "Document file"
// @Success 200 {object} dto.DataResponse{data=dto.DocsResponse}
// @Router /docs [post]
func (inst *Document) AddDocument(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		utils.CaseError(ctx, err)
		return
	}

	metaStr, ok := form.Value["meta"]
	if !ok {
		utils.CaseError(ctx, err)
		return
	}

	meta := &dto.Meta{}
	if err := json.Unmarshal([]byte(metaStr[0]), meta); err != nil {
		utils.CaseError(ctx, err)
		return
	}

	var jsonData map[string]any
	if jsonValues := form.Value["json"]; len(jsonValues) > 0 {
		if err := json.Unmarshal([]byte(jsonValues[0]), &jsonData); err != nil {
			utils.CaseError(ctx, err)
			return
		}
	}

	if err := inst.docService.AddDocument(ctx, &model.Document{
		Name:   meta.Name,
		Mime:   meta.Mime,
		File:   meta.File,
		Public: meta.Public,
		Grant:  meta.Grant,
	}, form.File["file"][0]); err != nil {
		utils.CaseError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.DataResponse{Data: dto.DocsResponse{
		JSON: jsonData,
		File: meta.Name,
	}})
}

// GetDocument godoc
// @Summary Get Documents
// @Description Get one document
// @Tags Document
// @Accept json
// @Accept mpfd
// @Produce json
// @Produce mpfd
// @Param uuid path string true "Document ID"
// @Param token query string true "docsorization token"
// @Success 200 {file} file "File content"
// @Success 200 {object} dto.DataResponse{data=dto.Meta} "File data"
// @Router /docs/{uuid} [get]
// @Router /docs/{uuid} [head]
func (inst *Document) GetDocument(ctx *gin.Context) {
	uuid := ctx.Param("uuid")
	if uuid == "" {
		utils.CaseError(ctx, utils.ErrorEmptyUUID)
		return
	}

	token := ctx.Query("token")
	if token == "" {
		utils.CaseError(ctx, utils.ErrorAuthFailed)
		return
	}

	document, err := inst.docService.GetDocument(ctx, uuid, token)
	if err != nil {
		utils.CaseError(ctx, err)
		return
	}

	if document.File {
		inst.sendFile(ctx, document)
		return
	}

	if ctx.Request.Method == http.MethodHead {
		ctx.Header("Content-Type", " application/json; charset=utf-8")
		ctx.Status(200)
		return
	}

	ctx.JSON(http.StatusOK, dto.DataResponse{
		Data: dto.Meta{
			ID:       document.UUID,
			Name:     document.Name,
			Mime:     document.Mime,
			File:     document.File,
			Public:   document.Public,
			CreateAt: document.CreateAt,
			Grant:    document.Grant,
		},
	})
}

// ListDocuments godoc
// @Summary List Documents
// @Description Get list of document
// @Tags Document
// @Accept json
// @Produce json
// @Param token query string true "docsorization token"
// @Param login query string false "Filter by grant login"
// @Param key query string false "Filter field key"
// @Param value query string false "Value of filter"
// @Param limit query string false "Limit, default 10"
// @Success 200 {file} file "File content"
// @Success 200 {object} dto.DataResponse{data=[]dto.Meta} "File data"
// @Router /docs [get]
// @Router /docs [head]
func (inst *Document) ListDocuments(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		utils.CaseError(ctx, utils.ErrorAuthFailed)
		return
	}

	listData := &model.DocumentFilterData{
		Login:        ctx.Query("login"),
		FiltredField: ctx.Query("key"),
		FiltredValue: ctx.Query("value"),
	}

	if err := inst.validateListData(ctx.Query("limit"), listData); err != nil {
		utils.CaseError(ctx, err)
		return
	}

	documents, err := inst.docService.ListDocuments(ctx, token, listData)
	if err != nil {
		utils.CaseError(ctx, err)
		return
	}

	if ctx.Request.Method == http.MethodHead {
		ctx.Header("Content-Type", " application/json; charset=utf-8")
		ctx.Status(200)
		return
	}

	ctx.JSON(http.StatusOK, dto.DataResponse{
		Data: inst.transformDocuments2Metas(documents),
	})
}

// DeleteDocument godoc
// @Summary Delete document Documents
// @Description Delete document by uuid
// @Tags Document
// @Accept json
// @Produce json
// @Param uuid path string true "Document ID"
// @Param token query string true "docsorization token"
// @Success 200 {file} file "File content"
// @Success 200 {object} dto.SuccessResponse{response=string} "File data"
// @Router /docs/{uuid} [delete]
func (inst *Document) DeleteDocument(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		utils.CaseError(ctx, utils.ErrorAuthFailed)
		return
	}

	uuid := ctx.Param("uuid")
	if uuid == "" {
		utils.CaseError(ctx, utils.ErrorEmptyUUID)
		return
	}

	if err := inst.docService.DeleteDocument(ctx, uuid, token); err != nil {
		utils.CaseError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, dto.SuccessResponse{Response: map[string]bool{
		token: true,
	}})

}

func (inst *Document) validateListData(limit string, listData *model.DocumentFilterData) error {
	if listData.FiltredField != "" && listData.FiltredValue == "" {
		return fmt.Errorf("%v: filtred value can't be null", utils.ErrorFilterFormat)
	}

	if listData.FiltredField == "" && listData.FiltredValue != "" {
		return fmt.Errorf("%v: missing filtred key", utils.ErrorFilterFormat)
	}

	if limit == "" {
		listData.Limit = 10
	}

	if limit != "" {
		var err error
		if listData.Limit, err = strconv.Atoi(limit); err != nil {
			return utils.ErrorLimitFormat
		}
	}

	return nil
}

func (inst *Document) transformDocuments2Metas(documents []model.Document) []dto.Meta {
	metas := make([]dto.Meta, 0)
	for _, document := range documents {
		metas = append(metas, dto.Meta{
			ID:       document.UUID,
			Name:     document.Name,
			Mime:     document.Mime,
			File:     document.File,
			Public:   document.Public,
			CreateAt: document.CreateAt,
			Grant:    document.Grant,
		})
	}
	return metas
}

func (inst *Document) sendFile(ctx *gin.Context, document *model.Document) {
	_, err := os.OpenFile(document.Path, os.O_WRONLY, 0666)
	if err != nil {
		// // TODO: delete file from db ?
		inst.log.Error("open file", zap.String("file", document.Path), zap.Error(err))

		ctx.JSON(http.StatusOK, dto.DataResponse{
			Data: dto.Meta{
				ID:       document.UUID,
				Name:     document.Name,
				Mime:     document.Mime,
				File:     document.File,
				Public:   document.Public,
				CreateAt: document.CreateAt,
				Grant:    document.Grant,
			},
		})
		return
	}
	ctx.Header("Content-Type", document.Mime)

	if ctx.Request.Method == http.MethodHead {
		ctx.Status(200)
		return
	}

	ctx.File(document.Path)
}
