package repositories

import (
	"context"
	"github.com/jmoiron/sqlx"
	"passkeeper/internal/models"
)

const (
	createPasswordSQL = "INSERT INTO t_pass (user_id, domen, username, password) VALUES (:user_id, :domen, :username, :password) RETURNING id"
	createCommentSQL  = "INSERT INTO t_comment (content_type, content_id, comment) VALUES (:content_type, :content_id, :comment) RETURNING id"
	updatePasswordSQL = "UPDATE t_pass SET domen = :domen, username = :username, password = :password, updated_at = :updated_at WHERE id = :id"
	updateCommentSQL  = "UPDATE t_comment SET comment = :comment, updated_at = :updated_at WHERE content_type = :content_type AND content_id = :content_id"

	getPasswordsByUserIDSQL = "SELECT tp.*, tc.comment FROM t_pass tp LEFT JOIN t_comment tc on tp.id = tc.content_id AND tc.content_type = $1 WHERE tp.user_id = $2;"
)

type PasswordRepository struct {
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db *sqlx.DB
	// storeCtx контекст, который отвечает за запросы
	ctx context.Context
}

func NewPasswordRepository(ctx context.Context, db SQLExecutor) *PasswordRepository {
	return &PasswordRepository{
		db:  db.(*sqlx.DB),
		ctx: ctx,
	}
}

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

func (pr *PasswordRepository) updatePasswordWithComment(tx *sqlx.Tx, password *models.PasswordContent, comment *models.Comment) error {
	_, err := tx.NamedExecContext(pr.ctx, updatePasswordSQL, password)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(pr.ctx, updateCommentSQL, comment)
	return err
}

func (pr *PasswordRepository) GetPasswordsByUserID(userID int64) ([]models.PasswordWithComment, error) {
	var passwords []models.PasswordWithComment
	err := pr.db.SelectContext(pr.ctx, &passwords, getPasswordsByUserIDSQL, models.TypePassword, userID)
	return passwords, err
}
