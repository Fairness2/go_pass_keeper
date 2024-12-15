package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"passkeeper/internal/models"
)

const (
	createPasswordSQL = "INSERT INTO t_pass (user_id, domen, username, password) VALUES (:user_id, :domen, :username, :password) RETURNING id"
	createCommentSQL  = "INSERT INTO t_comment (content_type, content_id, comment) VALUES (:content_type, :content_id, :comment) RETURNING id"
	updatePasswordSQL = "UPDATE t_pass SET domen = :domen, username = :username, password = :password, updated_at = :updated_at WHERE id = :id"
	updateCommentSQL  = "UPDATE t_comment SET comment = :comment, updated_at = :updated_at WHERE content_type = :content_type AND content_id = :content_id"

	getPasswordsByUserIDSQL      = "SELECT tp.*, tc.comment FROM t_pass tp LEFT JOIN t_comment tc on tp.id = tc.content_id AND tc.content_type = $1 WHERE tp.user_id = $2;"
	getPasswordsByUserIDAndIDSQL = "SELECT tp.* FROM t_pass tp WHERE tp.id = $1 AND tp.user_id = $2;"

	deletePasswordByUserIDAndIDSQL = "DELETE FROM t_pass WHERE id = $1 AND user_id = $2;"
	deleteCommentByContentID       = "DELETE FROM t_comment WHERE content_type = $1 AND content_id = $2;"
)

var (
	ErrNotExist = errors.New("not exist")
)

// PasswordRepository представляет собой хранилище для управления сущностями паролей и связанными с ними комментариями в базе данных.
// Он использует пул соединений с базой данных SQL и контекст для обработки запросов.
type PasswordRepository struct {
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db *sqlx.DB
	// storeCtx контекст, который отвечает за запросы
	ctx context.Context
}

// NewPasswordRepository создает и возвращает новый экземпляр PasswordRepository с предоставленным контекстом и SQLExecutor.
func NewPasswordRepository(ctx context.Context, db SQLExecutor) *PasswordRepository {
	return &PasswordRepository{
		db:  db.(*sqlx.DB),
		ctx: ctx,
	}
}

// Create вставляет или обновляет пароль вместе с соответствующим комментарием в транзакции базы данных.
func (pr *PasswordRepository) Create(password *models.PasswordContent, comment *models.Comment) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if password.ID != 0 {
		err = pr.updatePasswordWithComment(tx, password, comment)
	} else {
		err = pr.createPasswordWithComment(tx, password, comment)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// createPasswordWithComment вставляет новый пароль и связанный с ним комментарий в базу данных внутри транзакции.
func (pr *PasswordRepository) createPasswordWithComment(tx *sqlx.Tx, password *models.PasswordContent, comment *models.Comment) error {
	smth, err := tx.PrepareNamed(createPasswordSQL)
	if err != nil {
		return err
	}
	row := smth.QueryRowxContext(pr.ctx, password)
	if err := row.Scan(&password.ID); err != nil {
		return err
	}

	comment.ContentID = password.ID
	_, err = tx.NamedExecContext(pr.ctx, createCommentSQL, comment)
	return err
}

// updatePasswordWithComment обновляет пароль и связанный с ним комментарий в базе данных в рамках транзакции.
func (pr *PasswordRepository) updatePasswordWithComment(tx *sqlx.Tx, password *models.PasswordContent, comment *models.Comment) error {
	_, err := tx.NamedExecContext(pr.ctx, updatePasswordSQL, password)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(pr.ctx, updateCommentSQL, comment)
	return err
}

// GetPasswordsByUserID извлекает все пароли, связанные с указанным идентификатором пользователя, включая их комментарии. Возвращает список или ошибку.
func (pr *PasswordRepository) GetPasswordsByUserID(userID int64) ([]models.PasswordWithComment, error) {
	var passwords []models.PasswordWithComment
	err := pr.db.SelectContext(pr.ctx, &passwords, getPasswordsByUserIDSQL, models.TypePassword, userID)
	return passwords, err
}

// GetPasswordsByUserIDAndId извлекает пароль по его идентификатору и идентификатору связанного пользователя. Возвращает пароль или ошибку.
func (pr *PasswordRepository) GetPasswordsByUserIDAndId(userID int64, id int64) (*models.PasswordContent, error) {
	var password *models.PasswordContent
	row := pr.db.QueryRowxContext(pr.ctx, getPasswordsByUserIDAndIDSQL, id, userID)
	err := row.Err()
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Join(ErrNotExist, err)
	}
	if err != nil {
		return nil, err
	}
	if err = row.StructScan(password); err != nil {
		return nil, err
	}
	return password, err
}

// DeletePasswordByUserIDAndID удаляет пароль и связанные с ним комментарии для данного идентификатора пользователя и идентификатора пароля.
// Он выполняет два запроса на удаление в рамках транзакции базы данных.
// Возвращает ошибку, если транзакция или запросы завершаются неудачно.
func (pr *PasswordRepository) DeletePasswordByUserIDAndID(userID int64, id int64) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(pr.ctx, deletePasswordByUserIDAndIDSQL, id, userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(pr.ctx, deleteCommentByContentID, models.TypePassword, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
