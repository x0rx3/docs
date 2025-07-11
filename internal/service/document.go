package service

import (
	"context"
	"docs/internal/model"
	"docs/internal/repository"
	"docs/internal/utils"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

const (
	DocKeyFormat  = "doc:%s"
	DocsKeyFormat = "docs:%s:%s:%s:%d" // login:field:value:limit

	TagDocFormat       = "doc:%s"
	TagUserFormat      = "user:%s"
	TagFileNameFormat  = "fileName:%s"
	TagMimeFormat      = "mime:%s"
	TagIsFileFormat    = "isFile:%t"
	TagUserLoginFormat = "userLogin:%s"
	TagFilterFormat    = "filter:%s:%v"
)

type Document struct {
	log         *zap.Logger
	cache       Cacher
	grantRepo   repository.GrantRepository
	sessionRepo repository.SessionRepository
	docsRepo    repository.DocumentRepository
	uploadPath  string
}

func NewDocument(log *zap.Logger, uploadPath string, grantRepo repository.GrantRepository, docsRepo repository.DocumentRepository, sessionRepo repository.SessionRepository, cache Cacher) *Document {
	return &Document{
		log:         log,
		docsRepo:    docsRepo,
		uploadPath:  uploadPath,
		sessionRepo: sessionRepo,
		grantRepo:   grantRepo,
		cache:       cache,
	}
}

func (inst *Document) AddDocument(ctx context.Context, document *model.Document, file *multipart.FileHeader) error {
	inst.fielDocument(document)

	if document.File {
		if file == nil {
			return utils.ErrorEmptyFile
		}

		if err := inst.saveFile(document.Name, file); err != nil {
			return err
		}
	}

	if err := inst.docsRepo.CreateDocsWithGrant(ctx, document); err != nil {
		os.Remove(inst.uploadPath + document.Name)
		return err
	}

	go inst.invalidateDocument(document)

	return nil
}

func (inst *Document) GetDocument(ctx context.Context, uuid, sessionUUID string) (*model.Document, error) {
	var err error
	session, err := inst.sessionRepo.GetSessionByUUID(ctx, sessionUUID)
	if err != nil {
		return nil, utils.ErrorAuthFailed
	}

	_, err = inst.grantRepo.GetGrantByLoginAndDocUUID(ctx, uuid, session.UserLogin)
	if err != nil {
		if errors.Is(err, utils.ErrorNotFound) {
			return nil, utils.ErrorNoAccess
		}
		return nil, err
	}

	document := inst.fetchDocumentFromCache(uuid)
	if document != nil {
		inst.log.Debug("fetch document from cache")
		return document, nil
	}

	inst.log.Debug("document not found in cache")

	document, err = inst.docsRepo.GetDocumentWithGrantByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}

	inst.cache.Put(
		fmt.Sprintf(DocKeyFormat, document.UUID),
		document,
		1*time.Minute,
		inst.documentTags(document),
	)

	return document, nil
}

func (inst *Document) ListDocuments(ctx context.Context, sessionUUID string, data *model.DocumentFilterData) ([]model.Document, error) {
	if _, err := inst.sessionRepo.GetSessionByUUID(ctx, sessionUUID); err != nil {
		return nil, utils.ErrorAuthFailed
	}

	documents := inst.fetchDocumentsFromCache(data)
	if documents != nil {
		inst.log.Debug("fetch document from cache")
		return documents, nil
	}

	inst.log.Debug("document not found in cache")

	documents, err := inst.docsRepo.ListDocuments(ctx, data)
	if err != nil {
		return nil, err
	}

	inst.cache.Put(
		fmt.Sprintf(DocsKeyFormat, data.Login, data.FiltredField, data.FiltredValue, data.Limit),
		documents,
		1*time.Minute,
		inst.documentsTags(documents, data),
	)

	return documents, nil
}

func (inst *Document) DeleteDocument(ctx context.Context, uuid, sessionUUID string) error {
	_, err := inst.sessionRepo.GetSessionByUUID(ctx, sessionUUID)
	if err != nil {
		inst.log.Error("get session", zap.String("uuid", sessionUUID), zap.Error(err))
		return utils.ErrorAuthFailed
	}

	_, err = inst.grantRepo.GetGrantByLoginAndDocUUID(ctx, uuid, sessionUUID)
	if err != nil {
		if errors.Is(err, utils.ErrorNotFound) {
			return utils.ErrorNoAccess
		}
		return err
	}

	document, err := inst.docsRepo.GetDocumentByUUID(ctx, uuid)
	if err != nil {
		inst.log.Error("get document from db", zap.String("uuid", uuid), zap.Error(err))
		return err
	}

	if err := inst.removeFile(document.Path); err != nil {
		inst.log.Error("remove file", zap.String("path", document.Path), zap.Error(err))
	}

	if err := inst.docsRepo.DeleteDocument(ctx, uuid); err != nil {
		inst.log.Error("delete document", zap.String("uuid", uuid), zap.Error(err))
		return err
	}

	go inst.invalidateDocument(document)

	return nil
}

func (inst *Document) fielDocument(doc *model.Document) {
	doc.CreateAt = time.Now()
	doc.UUID = uuid.NewString()
	doc.Path = inst.uploadPath + "/" + filepath.Base(doc.Name)
}

func (inst *Document) saveFile(name string, file *multipart.FileHeader) error {
	if file == nil {
		return nil
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	if err = os.MkdirAll(filepath.Dir(inst.uploadPath), 0750); err != nil {
		return err
	}

	out, err := os.Create(inst.uploadPath + "/" + filepath.Base(name))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	if err != nil {
		return err
	}

	return nil
}

func (inst *Document) removeFile(path string) error {
	return os.Remove(path)
}

func (inst *Document) fetchDocumentFromCache(uuid string) *model.Document {
	value, exists := inst.cache.Get(
		fmt.Sprintf(DocKeyFormat, uuid),
	)
	if !exists {
		return nil
	}

	if document, ok := value.(*model.Document); ok {
		return document
	}

	inst.log.Error("unxpected model in cache", zap.String("key", fmt.Sprintf("docs:%s", uuid)))
	return nil
}

func (inst *Document) fetchDocumentsFromCache(data *model.DocumentFilterData) []model.Document {
	value, exists := inst.cache.Get(
		fmt.Sprintf(
			DocsKeyFormat, data.Login, data.FiltredField, data.FiltredValue, data.Limit,
		),
	)
	if !exists {
		return nil
	}

	if documents, ok := value.([]model.Document); ok {
		return documents
	}

	inst.log.Error("unxpected model in cache")
	return nil
}

func (inst *Document) invalidateDocument(document *model.Document) {
	inst.cache.InvalidateByTags(inst.documentTags(document))
	inst.cache.CleanExpired()
}

func (inst *Document) documentTags(document *model.Document) []string {
	tags := []string{
		fmt.Sprintf(TagFileNameFormat, document.Name),
		fmt.Sprintf(TagMimeFormat, document.Mime),
		fmt.Sprintf(TagIsFileFormat, document.File),
	}
	for _, grant := range document.Grant {
		tags = append(
			tags,
			fmt.Sprintf(TagUserLoginFormat, grant),
		)
	}

	return tags
}

func (inst *Document) documentsTags(documents []model.Document, listData *model.DocumentFilterData) []string {
	tags := []string{
		fmt.Sprintf("filter:%s:%v", listData.FiltredField, listData.FiltredValue),
	}

	for _, document := range documents {
		tags = append(tags,
			fmt.Sprintf(TagFileNameFormat, document.Name),
			fmt.Sprintf(TagMimeFormat, document.Mime),
			fmt.Sprintf(TagIsFileFormat, document.File),
		)
		for _, grant := range document.Grant {
			tags = append(
				tags,
				fmt.Sprintf(TagUserLoginFormat, grant),
			)
		}
	}
	return tags
}
