package repositories

import (
	"context"
	"github.com/jmoiron/sqlx"
	"passkeeper/internal/models"
)

const (
	createCardSQL = "INSERT INTO t_card (user_id, number, date, owner, cvv) VALUES (:user_id, :number, :date, :owner, :cvv) RETURNING id"
	updateCardSQL = "UPDATE t_card SET number = :number, date = :date, owner = :owner, cvv = :cvv, updated_at = :updated_at WHERE id = :id"

	getCardsByUserIDSQL      = "SELECT tp.*, tc.comment FROM t_card tp LEFT JOIN t_comment tc on tp.id = tc.content_id AND tc.content_type = $1 WHERE tp.user_id = $2;"
	getCardsByUserIDAndIDSQL = "SELECT tp.* FROM t_card tp WHERE tp.id = $1 AND tp.user_id = $2;"

	deleteCardByUserIDAndIDSQL = "DELETE FROM t_card WHERE id = $1 AND user_id = $2;"
)

// CardSQLSet представляет собой набор SQL-запросов для управления карточками и связанными с ними комментариями в базе данных.
var CardSQLSet = SQLSet{
	CreateContent:              createCardSQL,
	CreateComment:              createCommentSQL,
	UpdateContent:              updateCardSQL,
	UpdateComment:              updateCommentSQL,
	GetContentByUserID:         getCardsByUserIDSQL,
	GetContentByUserIDAndID:    getCardsByUserIDAndIDSQL,
	DeleteContentByUserIDAndID: deleteCardByUserIDAndIDSQL,
	DeleteCommentByContentID:   deleteCommentByContentID,
}

// NewCardRepository инициализирует новый CrudRepository для CardContent и CardWithComment, используя предоставленный контекст и базу данных.
func NewCardRepository(ctx context.Context, db SQLExecutor) *CrudRepository[models.CardContent, models.CardWithComment] {
	return &CrudRepository[models.CardContent, models.CardWithComment]{
		db:          db.(*sqlx.DB),
		ctx:         ctx,
		sqlSet:      CardSQLSet,
		typeContent: models.TypeCard,
	}
}
