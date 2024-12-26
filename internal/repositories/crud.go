package repositories

import (
	"context"
	"database/sql"
	"errors"
	"passkeeper/internal/models"
)

// SQLSet представляет собой набор SQL-запросов для управления содержимым и соответствующими комментариями в базе данных.
type SQLSet struct {
	CreateContent              string // SQL-запрос для создания записи контента.
	CreateComment              string // SQL-запрос для создания записи комментария, связанной с контентом.
	UpdateContent              string // SQL-запрос для обновления определенной записи контента.
	UpdateComment              string // SQL для обновления записи комментария для определенного контента.
	GetContentByUserID         string // SQL-запрос для получения нескольких записей контента по идентификатору пользователя.
	GetContentByUserIDAndID    string // SQL-запрос для получения определенной записи контента по идентификатору пользователя и идентификатору контента.
	DeleteContentByUserIDAndID string // SQL-запрос для удаления определенной записи контента по идентификатору пользователя и идентификатору контента.
	DeleteCommentByContentID   string // SQL-запрос для удаления комментариев, связанных с определенным контентом.
}

// Contentable представляет сущность, которая предоставляет уникальный идентификатор с помощью метода GetID.
type Contentable interface {
	GetID() int64
}

// CrudRepository представляет собой хранилище для управления контентными сущностями и связанными с ними комментариями в базе данных.
// Он использует пул соединений с базой данных SQL и контекст для обработки запросов.
type CrudRepository[T Contentable, Y Contentable] struct {
	db          SQLExecutor        // Пул соединений с базой данных, которыми может пользоваться хранилище
	sqlSet      SQLSet             // Набор SQL запросов для контента
	typeContent models.ContentType // Тип контента
}

// Create вставляет или обновляет контент вместе с соответствующим комментарием в транзакции базы данных.
func (pr *CrudRepository[T, Y]) Create(ctx context.Context, content T, comment models.Comment) error {
	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if content.GetID() != 0 {
		err = pr.updateWithComment(ctx, tx, content, comment)
	} else {
		err = pr.createWithComment(ctx, tx, content, comment)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// createWithComment вставляет новый контент и связанный с ним комментарий в базу данных внутри транзакции.
func (pr *CrudRepository[T, Y]) createWithComment(ctx context.Context, tx ITX, content T, comment models.Comment) error {
	smth, err := tx.PrepareNamed(pr.sqlSet.CreateContent)
	if err != nil {
		return err
	}

	row := smth.QueryRowxContext(ctx, content)
	var id int64
	if err := row.Scan(&id); err != nil {
		return err
	}

	comment.ContentID = id
	_, err = tx.NamedExecContext(ctx, pr.sqlSet.CreateComment, comment)
	return err
}

// updateWithComment обновляет информацию о контенте и связанный с ним комментарий в базе данных в рамках транзакции.
func (pr *CrudRepository[T, Y]) updateWithComment(ctx context.Context, tx ITX, content T, comment models.Comment) error {
	_, err := tx.NamedExecContext(ctx, pr.sqlSet.UpdateContent, content)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(ctx, pr.sqlSet.UpdateComment, comment)
	return err
}

// GetByUserID извлекает все информации о контенте, связанные с указанным идентификатором пользователя, включая их комментарии. Возвращает список или ошибку.
func (pr *CrudRepository[T, Y]) GetByUserID(ctx context.Context, userID int64) ([]Y, error) {
	var result []Y
	err := pr.db.SelectContext(ctx, &result, pr.sqlSet.GetContentByUserID, pr.typeContent, userID)
	return result, err
}

// GetByUserIDAndId извлекает информацию о контенте по его идентификатору и идентификатору связанного пользователя. Возвращает текст или ошибку.
func (pr *CrudRepository[T, Y]) GetByUserIDAndId(ctx context.Context, userID int64, id int64) (*T, error) {
	row := pr.db.QueryRowxContext(ctx, pr.sqlSet.GetContentByUserIDAndID, id, userID)
	err := row.Err()
	if err != nil {
		return nil, err
	}
	var result T
	if err = row.StructScan(&result); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(ErrNotExist, err)
		}
		return nil, err
	}
	return &result, nil
}

// DeleteByUserIDAndID удаляет информацию о контенте и связанные с ним комментарии для данного идентификатора пользователя и идентификатора текста.
// Он выполняет два запроса на удаление в рамках транзакции базы данных.
// Возвращает ошибку, если транзакция или запросы завершаются неудачно.
func (pr *CrudRepository[T, Y]) DeleteByUserIDAndID(ctx context.Context, userID int64, id int64) error {
	tx, err := pr.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, pr.sqlSet.DeleteContentByUserIDAndID, id, userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, pr.sqlSet.DeleteCommentByContentID, pr.typeContent, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
