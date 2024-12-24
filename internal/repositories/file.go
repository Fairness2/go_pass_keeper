package repositories

import (
	"passkeeper/internal/models"
)

const (
	createFileSQL = "INSERT INTO t_file (user_id, name, file_path) VALUES (:user_id, :name, :file_path) RETURNING id"
	updateFileSQL = "UPDATE t_file SET name = :name, updated_at = :updated_at WHERE id = :id"

	getFileByUserIDSQL      = "SELECT tt.*, tc.comment FROM t_file tt LEFT JOIN t_comment tc on tt.id = tc.content_id AND tc.content_type = $1 WHERE tt.user_id = $2;"
	getFileByUserIDAndIDSQL = "SELECT tt.* FROM t_file tt WHERE tt.id = $1 AND tt.user_id = $2;"

	deleteFileByUserIDAndIDSQL = "DELETE FROM t_file WHERE id = $1 AND user_id = $2;"
)

// FileSQLSet содержит шаблоны запросов SQL для обработки содержимого файлов и связанных комментариев в базе данных.
var FileSQLSet = SQLSet{
	CreateContent:              createFileSQL,
	CreateComment:              createCommentSQL,
	UpdateContent:              updateFileSQL,
	UpdateComment:              updateCommentSQL,
	GetContentByUserID:         getFileByUserIDSQL,
	GetContentByUserIDAndID:    getFileByUserIDAndIDSQL,
	DeleteContentByUserIDAndID: deleteFileByUserIDAndIDSQL,
	DeleteCommentByContentID:   deleteCommentByContentID,
}

// NewFileRepository создает новый экземпляр CrudRepository для управления сущностями FileContent и FileWithComment в базе данных.
func NewFileRepository(db SQLExecutor) *CrudRepository[models.FileContent, models.FileWithComment] {
	return &CrudRepository[models.FileContent, models.FileWithComment]{
		db:          db,
		sqlSet:      FileSQLSet,
		typeContent: models.TypeFile,
	}
}
