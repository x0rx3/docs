package postgres

import (
	"context"
	"docs/internal/model"
	"docs/internal/utils"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Document struct {
	log  *zap.Logger
	pool *pgxpool.Pool
}

func NewDocument(log *zap.Logger, pool *pgxpool.Pool) *Document {
	return &Document{log, pool}
}

func (inst *Document) CreateDocsWithGrant(ctx context.Context, document *model.Document) error {
	tx, err := inst.pool.Begin(ctx)
	if err != nil {
		return err
	}

	if err := inst.insertDocument(tx, ctx, document); err != nil {
		tx.Rollback(ctx)
		return err
	}

	if err := inst.insertGrant(tx, ctx, document.UUID, document.Grant); err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (inst *Document) GetDocumentWithGrantByUUID(ctx context.Context, uuid string) (*model.Document, error) {
	var document *model.Document
	rows, err := inst.pool.Query(ctx, inst.selectWithGrantQuery(), uuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrorNotFound
		}
		return nil, err
	}

	for rows.Next() {
		var (
			uuid      string
			name      string
			mime      string
			file      bool
			public    bool
			createAt  time.Time
			path      string
			userLogin *string
		)

		if err := rows.Scan(&uuid, &name, &mime, &file, &public, &createAt, &path, &userLogin); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}

		if document == nil {
			document = &model.Document{
				UUID:     uuid,
				Name:     name,
				Mime:     mime,
				File:     file,
				Public:   public,
				CreateAt: createAt,
				Path:     path,
			}
		}

		if userLogin != nil {
			document.Grant = append(document.Grant, *userLogin)
		}
	}

	if document == nil {
		return nil, utils.ErrorNotFound
	}

	return document, nil
}

func (inst *Document) GetDocumentByUUID(ctx context.Context, uuid string) (*model.Document, error) {
	return inst.selectDocument(ctx, uuid)
}

func (inst *Document) ListDocuments(ctx context.Context, data *model.DocumentFilterData) ([]model.Document, error) {

	sql, filterValues, err := inst.listFilters(inst.selectWithLimitQuery(), data)
	if err != nil {
		return nil, err
	}

	inst.log.Debug("select sql", zap.String("sql", sql))

	rows, err := inst.pool.Query(ctx, sql, filterValues...)
	if err != nil {
		return nil, err
	}

	documents := make([]model.Document, 0)
	for rows.Next() {
		document := &model.Document{}
		if err := rows.Scan(
			&document.UUID,
			&document.Name,
			&document.Mime,
			&document.File,
			&document.Public,
			&document.CreateAt,
			&document.Path,
			&document.Grant,
		); err != nil {
			return nil, err
		}

		documents = append(documents, *document)
	}

	return documents, nil
}

func (inst *Document) DeleteDocument(ctx context.Context, uuid string) error {
	if _, err := inst.pool.Exec(ctx, `DELETE FROM documents WHERE uuid = $1`, uuid); err != nil {
		return err
	}

	return nil
}

func (inst *Document) selectDocument(ctx context.Context, uuid string) (*model.Document, error) {
	sql := `SELECT 
		documents.uuid,
		documents.name,
		documents.mime,
		documents.file,
		documents.public,
		documents.create_at,
		documents.path
	FROM documents WHERE uuid = $1;
	`
	document := &model.Document{}

	inst.log.Debug("select sql", zap.String("sql", sql))

	if err := inst.pool.QueryRow(ctx, sql, uuid).Scan(
		&document.UUID,
		&document.Name,
		&document.Mime,
		&document.File,
		&document.Public,
		&document.CreateAt,
		&document.Path,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.ErrorNotFound
		}

		return nil, err
	}

	return document, nil

}

func (inst *Document) insertDocument(tx pgx.Tx, ctx context.Context, document *model.Document) error {
	if _, err := tx.Exec(
		ctx,
		`INSERT INTO documents
		(uuid, name, mime, file, public, create_at, path)
		VALUES ($1, $2, $3, $4, $5, $6, $7);`,
		document.UUID,
		document.Name,
		document.Mime,
		document.File,
		document.Public,
		document.CreateAt,
		document.Path,
	); err != nil {
		return err
	}
	return nil
}

func (inst *Document) insertGrant(tx pgx.Tx, ctx context.Context, documentUUID string, grant []string) error {
	const errorForiengKeyCode = "23503"

	sql, values := inst.buildInsertGrantQuery(documentUUID, grant)

	inst.log.Debug("insert sql", zap.String("sql", sql))

	_, err := tx.Exec(ctx, sql, values...)
	if err != nil {
		tx.Rollback(ctx)
		if pgerr, ok := err.(*pgconn.PgError); ok && pgerr.Code == errorForiengKeyCode {
			return utils.ErrorInvalidGrant
		}
		return err
	}

	return nil
}

func (inst *Document) buildInsertGrantQuery(documentUUID string, grant []string) (string, []any) {
	sql := `INSERT INTO document_grants (document_uuid, user_login) VALUES %s;`

	placeholder := `($%d, $%d)`
	placeholders := make([]string, 0)
	values := make([]any, 0)
	sum := 1

	for _, grant := range grant {
		placeholders = append(placeholders, fmt.Sprintf(placeholder, sum, sum+1))
		values = append(values, documentUUID, grant)
		sum = sum + 2
	}

	sql = fmt.Sprintf(sql, strings.Join(placeholders, ", "))

	return sql, values
}

func (inst *Document) listFilters(sql string, data *model.DocumentFilterData) (string, []any, error) {
	numFilter := 1
	filterPlaceholders := make([]string, 0)
	filterValues := make([]any, 0)

	if data.Login != "" {
		filterPlaceholders = append(
			filterPlaceholders,
			fmt.Sprintf("(%s = $%d)", "document_grants.user_login", numFilter),
		)
		filterValues = append(filterValues, data.Login)
		numFilter++
	}

	if data.FiltredField != "" {
		filterWhiteList := map[string]string{
			"name":      "documents.name",
			"id":        "documents.uuid",
			"mime":      "documents.mime",
			"file":      "documents.file",
			"publuc":    "documents.public",
			"create_at": "documents.create_at",
		}

		fieldDBName, ok := filterWhiteList[data.FiltredField]
		if !ok {
			return "", nil, utils.ErrorFilterFormat
		}

		if numFilter > 1 {
			filterPlaceholders = append(
				filterPlaceholders,
				fmt.Sprintf("(%s = $%d)", fieldDBName, numFilter),
			)
			filterValues = append(filterValues, data.FiltredValue)
		} else {
			filterPlaceholders = append(
				filterPlaceholders,
				fmt.Sprintf("(%s = $%d)", fieldDBName, numFilter),
			)
			filterValues = append(filterValues, data.FiltredValue)
		}
		numFilter++
	}

	if numFilter == 1 {
		sql = fmt.Sprintf(sql, "", fmt.Sprintf("LIMIT %d", data.Limit))
	} else {
		sql = fmt.Sprintf(
			sql,
			fmt.Sprintf("WHERE %s", strings.Join(filterPlaceholders, " AND ")),
			fmt.Sprintf("LIMIT %d", data.Limit),
		)
	}

	return sql, filterValues, nil
}

func (inst *Document) selectWithLimitQuery() string {
	return `SELECT 
		documents.uuid,
		documents.name,
		documents.mime,
		documents.file,
		documents.public,
		documents.create_at,
		documents.path,
		array_remove(array_agg(document_grants.user_login), NULL)
	FROM documents
	LEFT JOIN document_grants ON documents.uuid = document_uuid 
	%s
	GROUP BY 
		documents.uuid,
		documents.name,
		documents.mime,
		documents.file,
		documents.public,
		documents.create_at,
		documents.path
	%s;`
}

func (inst *Document) selectWithGrantQuery() string {
	return `
	SELECT 
		documents.uuid,
		documents.name,
		documents.mime,
		documents.file,
		documents.public,
		documents.create_at,
		documents.path,
		document_grants.user_login
	from documents
	LEFT JOIN document_grants ON documents.uuid = document_uuid
	WHERE documents.uuid = $1;
	`
}
