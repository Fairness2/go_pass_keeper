package repositories

import (
	"errors"
	"passkeeper/internal/models"
)

const (
	createPasswordSQL = "INSERT INTO t_pass (user_id, domen, username, password) VALUES (:user_id, :domen, :username, :password) RETURNING id"
	createCommentSQL  = "INSERT INTO t_comment (content_type, content_id, comment) VALUES (:content_type, :content_id, :comment) RETURNING id"
	updatePasswordSQL = "UPDATE t_pass SET domen = :domen, username = :username, password = :password, updated_at = :updated_at WHERE id = :id"
	updateCommentSQL  = "UPDATE t_comment SET comment = :comment, updated_at = :updated_at WHERE content_type = :content_type AND content_id = :content_id" // TODO дата сохранения в бд

	getPasswordsByUserIDSQL      = "SELECT tp.*, tc.comment FROM t_pass tp LEFT JOIN t_comment tc on tp.id = tc.content_id AND tc.content_type = $1 WHERE tp.user_id = $2;"
	getPasswordsByUserIDAndIDSQL = "SELECT tp.* FROM t_pass tp WHERE tp.id = $1 AND tp.user_id = $2;"

	deletePasswordByUserIDAndIDSQL = "DELETE FROM t_pass WHERE id = $1 AND user_id = $2;"
	deleteCommentByContentID       = "DELETE FROM t_comment WHERE content_type = $1 AND content_id = $2;"
)

var (
	ErrNotExist = errors.New("not exist") // Ошибка, что файл не существует
)

// PasswordSQLSet — это предварительно настроенный набор SQLSet, используемый для обработки операций, связанных с содержимым пароля и связанными с ним комментариями.
var PasswordSQLSet = SQLSet{
	CreateContent:              createPasswordSQL,
	CreateComment:              createCommentSQL,
	UpdateContent:              updatePasswordSQL,
	UpdateComment:              updateCommentSQL,
	GetContentByUserID:         getPasswordsByUserIDSQL,
	GetContentByUserIDAndID:    getPasswordsByUserIDAndIDSQL,
	DeleteContentByUserIDAndID: deletePasswordByUserIDAndIDSQL,
	DeleteCommentByContentID:   deleteCommentByContentID,
}

// NewPasswordRepository создает и возвращает новый экземпляр CrudRepository для паролей с предоставленным контекстом и SQLExecutor
func NewPasswordRepository(db SQLExecutor) *CrudRepository[models.PasswordContent, models.PasswordWithComment] {
	return &CrudRepository[models.PasswordContent, models.PasswordWithComment]{
		db:          db,
		sqlSet:      PasswordSQLSet,
		typeContent: models.TypePassword,
	}
}
