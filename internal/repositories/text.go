package repositories

import (
	"passkeeper/internal/models"
)

const (
	createTextSQL = "INSERT INTO t_text (user_id, text_data) VALUES (:user_id, :text_data) RETURNING id"
	updateTextSQL = "UPDATE t_text SET text_data = :text_data, updated_at = :updated_at WHERE id = :id"

	getTextByUserIDSQL      = "SELECT tt.*, tc.comment FROM t_text tt LEFT JOIN t_comment tc on tt.id = tc.content_id AND tc.content_type = $1 WHERE tt.user_id = $2;"
	getTextByUserIDAndIDSQL = "SELECT tt.* FROM t_text tt WHERE tt.id = $1 AND tt.user_id = $2;"

	deleteTextByUserIDAndIDSQL = "DELETE FROM t_text WHERE id = $1 AND user_id = $2;"
)

// TextSQLSet — это предопределенный набор SQLSet для управления текстовым содержимым и соответствующими комментариями в базе данных.
var TextSQLSet = SQLSet{
	CreateContent:              createTextSQL,
	CreateComment:              createCommentSQL,
	UpdateContent:              updateTextSQL,
	UpdateComment:              updateCommentSQL,
	GetContentByUserID:         getTextByUserIDSQL,
	GetContentByUserIDAndID:    getTextByUserIDAndIDSQL,
	DeleteContentByUserIDAndID: deleteTextByUserIDAndIDSQL,
	DeleteCommentByContentID:   deleteCommentByContentID,
}

// NewTextRepository создает новый экземпляр CrudRepository для управления текстовым содержимым и связанными комментариями в базе данных.
func NewTextRepository(db SQLExecutor) *CrudRepository[models.TextContent, models.TextWithComment] {
	return &CrudRepository[models.TextContent, models.TextWithComment]{
		db:          db,
		sqlSet:      TextSQLSet,
		typeContent: models.TypeText,
	}
}
