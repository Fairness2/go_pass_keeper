package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"passkeeper/internal/models"
)

type SQLSet struct {
	CreateContent              string
	CreateComment              string
	UpdateContent              string
	UpdateComment              string
	GetContentByUserID         string
	GetContentByUserIDAndID    string
	DeleteContentByUserIDAndID string
	DeleteCommentByContentID   string
}

type Contentable interface {
	GetID() int64
	SetID(int64)
}

// CrudRepository представляет собой хранилище для управления контентными сущностями и связанными с ними комментариями в базе данных.
// Он использует пул соединений с базой данных SQL и контекст для обработки запросов.
type CrudRepository struct {
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db *sqlx.DB
	// storeCtx контекст, который отвечает за запросы
	ctx    context.Context
	sqlSet SQLSet
}

// NewCrudRepository создает и возвращает новый экземпляр FileRepository с предоставленным контекстом и SQLExecutor.
func NewCrudRepository(ctx context.Context, db SQLExecutor, sqlSet SQLSet) *CrudRepository {
	return &CrudRepository{
		db:     db.(*sqlx.DB),
		ctx:    ctx,
		sqlSet: sqlSet,
	}
}

// Create вставляет или обновляет контент вместе с соответствующим комментарием в транзакции базы данных.
func (pr *CrudRepository) Create(content Contentable, comment *models.Comment) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if content.GetID() != 0 {
		err = pr.updateWithComment(tx, content, comment)
	} else {
		err = pr.createWithComment(tx, content, comment)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// createWithComment вставляет новый контент и связанный с ним комментарий в базу данных внутри транзакции.
func (pr *CrudRepository) createWithComment(tx *sqlx.Tx, content Contentable, comment *models.Comment) error {
	smth, err := tx.PrepareNamed(pr.sqlSet.CreateContent)
	if err != nil {
		return err
	}
	row := smth.QueryRowxContext(pr.ctx, content)
	var id int64
	if err := row.Scan(&id); err != nil {
		return err
	}

	comment.ContentID = id
	_, err = tx.NamedExecContext(pr.ctx, pr.sqlSet.CreateComment, comment)
	return err
}

// updateWithComment обновляет информацию о файле и связанный с ним комментарий в базе данных в рамках транзакции.
func (pr *CrudRepository) updateWithComment(tx *sqlx.Tx, content Contentable, comment *models.Comment) error {
	_, err := tx.NamedExecContext(pr.ctx, pr.sqlSet.UpdateContent, content)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(pr.ctx, pr.sqlSet.UpdateComment, comment)
	return err
}

// GetByUserID извлекает все информации офайлах, связанные с указанным идентификатором пользователя, включая их комментарии. Возвращает список или ошибку.
func (pr *CrudRepository) GetByUserID(userID int64, result *[]Contentable, typeContent models.ContentType) error {
	err := pr.db.SelectContext(pr.ctx, result, pr.sqlSet.GetContentByUserID, typeContent, userID)
	return err
}

// GetByUserIDAndId извлекает информацию о файле по его идентификатору и идентификатору связанного пользователя. Возвращает текст или ошибку.
func (pr *CrudRepository) GetByUserIDAndId(userID int64, id int64, result *Contentable) error {
	row := pr.db.QueryRowxContext(pr.ctx, pr.sqlSet.GetContentByUserIDAndID, id, userID)
	err := row.Err()
	if err != nil {
		return err
	}
	if err = row.StructScan(result); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Join(ErrNotExist, err)
		}
		return err
	}
	return err
}

// DeleteByUserIDAndID удаляет информацию о файле и связанные с ним комментарии для данного идентификатора пользователя и идентификатора текста.
// Он выполняет два запроса на удаление в рамках транзакции базы данных.
// Возвращает ошибку, если транзакция или запросы завершаются неудачно.
func (pr *CrudRepository) DeleteByUserIDAndID(userID int64, id int64, typeContent models.ContentType) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(pr.ctx, pr.sqlSet.DeleteContentByUserIDAndID, id, userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(pr.ctx, pr.sqlSet.DeleteCommentByContentID, typeContent, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
