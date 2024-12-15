package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"passkeeper/internal/models"
)

const (
	createTextSQL = "INSERT INTO t_text (user_id, text_data) VALUES (:user_id, :text_data) RETURNING id"
	updateTextSQL = "UPDATE t_text SET text_data = :text_data, updated_at = :updated_at WHERE id = :id"

	getTextByUserIDSQL      = "SELECT tt.*, tc.comment FROM t_text tt LEFT JOIN t_comment tc on tt.id = tc.content_id AND tc.content_type = $1 WHERE tt.user_id = $2;"
	getTextByUserIDAndIDSQL = "SELECT tt.* FROM t_text tt WHERE tt.id = $1 AND tt.user_id = $2;"

	deleteTextByUserIDAndIDSQL = "DELETE FROM t_text WHERE id = $1 AND user_id = $2;"
)

// TextRepository представляет собой хранилище для управления сущностями произвольного текста и связанными с ними комментариями в базе данных.
// Он использует пул соединений с базой данных SQL и контекст для обработки запросов.
type TextRepository struct {
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db *sqlx.DB
	// storeCtx контекст, который отвечает за запросы
	ctx context.Context
}

// NewTextRepository создает и возвращает новый экземпляр TextRepository с предоставленным контекстом и SQLExecutor.
func NewTextRepository(ctx context.Context, db SQLExecutor) *TextRepository {
	return &TextRepository{
		db:  db.(*sqlx.DB),
		ctx: ctx,
	}
}

// Create вставляет или обновляет текст вместе с соответствующим комментарием в транзакции базы данных.
func (pr *TextRepository) Create(text *models.TextContent, comment *models.Comment) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if text.ID != 0 {
		err = pr.updateTextWithComment(tx, text, comment)
	} else {
		err = pr.createTextWithComment(tx, text, comment)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// createTextWithComment вставляет новый текст и связанный с ним комментарий в базу данных внутри транзакции.
func (pr *TextRepository) createTextWithComment(tx *sqlx.Tx, text *models.TextContent, comment *models.Comment) error {
	smth, err := tx.PrepareNamed(createTextSQL)
	if err != nil {
		return err
	}
	row := smth.QueryRowxContext(pr.ctx, text)
	if err := row.Scan(&text.ID); err != nil {
		return err
	}

	comment.ContentID = text.ID
	_, err = tx.NamedExecContext(pr.ctx, createCommentSQL, comment)
	return err
}

// updateTextWithComment обновляет текст и связанный с ним комментарий в базе данных в рамках транзакции.
func (pr *TextRepository) updateTextWithComment(tx *sqlx.Tx, text *models.TextContent, comment *models.Comment) error {
	_, err := tx.NamedExecContext(pr.ctx, updateTextSQL, text)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(pr.ctx, updateCommentSQL, comment)
	return err
}

// GetTextsByUserID извлекает все Тексты, связанные с указанным идентификатором пользователя, включая их комментарии. Возвращает список или ошибку.
func (pr *TextRepository) GetTextsByUserID(userID int64) ([]models.TextWithComment, error) {
	var texts []models.TextWithComment
	err := pr.db.SelectContext(pr.ctx, &texts, getTextByUserIDSQL, models.TypeText, userID)
	return texts, err
}

// GetTextByUserIDAndId извлекает текст по его идентификатору и идентификатору связанного пользователя. Возвращает текст или ошибку.
func (pr *TextRepository) GetTextByUserIDAndId(userID int64, id int64) (*models.TextContent, error) {
	var text *models.TextContent
	row := pr.db.QueryRowxContext(pr.ctx, getTextByUserIDAndIDSQL, id, userID)
	err := row.Err()
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Join(ErrNotExist, err)
	}
	if err != nil {
		return nil, err
	}
	if err = row.StructScan(text); err != nil {
		return nil, err
	}
	return text, err
}

// DeleteTextByUserIDAndID удаляет текст и связанные с ним комментарии для данного идентификатора пользователя и идентификатора текста.
// Он выполняет два запроса на удаление в рамках транзакции базы данных.
// Возвращает ошибку, если транзакция или запросы завершаются неудачно.
func (pr *TextRepository) DeleteTextByUserIDAndID(userID int64, id int64) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(pr.ctx, deleteTextByUserIDAndIDSQL, id, userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(pr.ctx, deleteCommentByContentID, models.TypeText, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
