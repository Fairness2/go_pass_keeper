package repositories

import (
	"context"
	"database/sql"
	"errors"
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

// CardRepository представляет собой хранилище для управления сущностями карт и связанными с ними комментариями в базе данных.
// Он использует пул соединений с базой данных SQL и контекст для обработки запросов.
type CardRepository struct {
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db *sqlx.DB
	// storeCtx контекст, который отвечает за запросы
	ctx context.Context
}

// NewCardRepository создает и возвращает новый экземпляр CardRepository с предоставленным контекстом и SQLExecutor.
func NewCardRepository(ctx context.Context, db SQLExecutor) *CardRepository {
	return &CardRepository{
		db:  db.(*sqlx.DB),
		ctx: ctx,
	}
}

// Create вставляет или обновляет карту вместе с соответствующим комментарием в транзакции базы данных.
func (pr *CardRepository) Create(card *models.CardContent, comment *models.Comment) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if card.ID != 0 {
		err = pr.updateCardWithComment(tx, card, comment)
	} else {
		err = pr.createCardWithComment(tx, card, comment)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// createCardWithComment вставляет новую карту и связанный с ней комментарий в базу данных внутри транзакции.
func (pr *CardRepository) createCardWithComment(tx *sqlx.Tx, card *models.CardContent, comment *models.Comment) error {
	smth, err := tx.PrepareNamed(createCardSQL)
	if err != nil {
		return err
	}
	row := smth.QueryRowxContext(pr.ctx, card)
	if err := row.Scan(&card.ID); err != nil {
		return err
	}

	comment.ContentID = card.ID
	_, err = tx.NamedExecContext(pr.ctx, createCommentSQL, comment)
	return err
}

// updateCardWithComment обновляет карту и связанный с ней комментарий в базе данных в рамках транзакции.
func (pr *CardRepository) updateCardWithComment(tx *sqlx.Tx, card *models.CardContent, comment *models.Comment) error {
	_, err := tx.NamedExecContext(pr.ctx, updateCardSQL, card)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(pr.ctx, updateCommentSQL, comment)
	return err
}

// GetCardsByUserID извлекает все карты, связанные с указанным идентификатором пользователя, включая их комментарии. Возвращает список или ошибку.
func (pr *CardRepository) GetCardsByUserID(userID int64) ([]models.CardWithComment, error) {
	var cards []models.CardWithComment
	err := pr.db.SelectContext(pr.ctx, &cards, getCardsByUserIDSQL, models.TypeCard, userID)
	return cards, err
}

// GetCardByUserIDAndId извлекает карту по его идентификатору и идентификатору связанного пользователя. Возвращает карту или ошибку.
func (pr *CardRepository) GetCardByUserIDAndId(userID int64, id int64) (*models.CardContent, error) {
	row := pr.db.QueryRowxContext(pr.ctx, getCardsByUserIDAndIDSQL, id, userID)
	err := row.Err()
	if err != nil {
		return nil, err
	}
	var card models.CardContent
	if err = row.StructScan(&card); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(ErrNotExist, err)
		}
		return nil, err
	}
	return &card, err
}

// DeleteCardByUserIDAndID удаляет карту и связанные с ним комментарии для данного идентификатора пользователя и идентификатора пароля.
// Он выполняет два запроса на удаление в рамках транзакции базы данных.
// Возвращает ошибку, если транзакция или запросы завершаются неудачно.
func (pr *CardRepository) DeleteCardByUserIDAndID(userID int64, id int64) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(pr.ctx, deleteCardByUserIDAndIDSQL, id, userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(pr.ctx, deleteCommentByContentID, models.TypeCard, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
