package repositories

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"passkeeper/internal/models"
)

const (
	createFileSQL = "INSERT INTO t_file (user_id, name, file_path) VALUES (:user_id, :name, :file_path) RETURNING id"
	updateFileSQL = "UPDATE t_file SET name = :name, updated_at = :updated_at WHERE id = :id"

	getFileByUserIDSQL      = "SELECT tt.*, tc.comment FROM t_file tt LEFT JOIN t_comment tc on tt.id = tc.content_id AND tc.content_type = $1 WHERE tt.user_id = $2;"
	getFileByUserIDAndIDSQL = "SELECT tt.* FROM t_file tt WHERE tt.id = $1 AND tt.user_id = $2;"

	deleteFileByUserIDAndIDSQL = "DELETE FROM t_file WHERE id = $1 AND user_id = $2;"
)

// FileRepository представляет собой хранилище для управления сущностями информации о файлах и связанными с ними комментариями в базе данных.
// Он использует пул соединений с базой данных SQL и контекст для обработки запросов.
type FileRepository struct {
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db *sqlx.DB
	// storeCtx контекст, который отвечает за запросы
	ctx context.Context
}

// NewFileRepository создает и возвращает новый экземпляр FileRepository с предоставленным контекстом и SQLExecutor.
func NewFileRepository(ctx context.Context, db SQLExecutor) *FileRepository {
	return &FileRepository{
		db:  db.(*sqlx.DB),
		ctx: ctx,
	}
}

// Create вставляет или обновляет информацию о файле вместе с соответствующим комментарием в транзакции базы данных.
func (pr *FileRepository) Create(file *models.FileContent, comment *models.Comment) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if file.ID != 0 {
		err = pr.updateFileWithComment(tx, file, comment)
	} else {
		err = pr.createFileWithComment(tx, file, comment)
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}

// createFileWithComment вставляет новую информацию о файле и связанный с ним комментарий в базу данных внутри транзакции.
func (pr *FileRepository) createFileWithComment(tx *sqlx.Tx, file *models.FileContent, comment *models.Comment) error {
	smth, err := tx.PrepareNamed(createFileSQL)
	if err != nil {
		return err
	}
	row := smth.QueryRowxContext(pr.ctx, file)
	if err := row.Scan(&file.ID); err != nil {
		return err
	}

	comment.ContentID = file.ID
	_, err = tx.NamedExecContext(pr.ctx, createCommentSQL, comment)
	return err
}

// updateFileWithComment обновляет информацию о файле и связанный с ним комментарий в базе данных в рамках транзакции.
func (pr *FileRepository) updateFileWithComment(tx *sqlx.Tx, file *models.FileContent, comment *models.Comment) error {
	_, err := tx.NamedExecContext(pr.ctx, updateFileSQL, file)
	if err != nil {
		return err
	}

	_, err = tx.NamedExecContext(pr.ctx, updateCommentSQL, comment)
	return err
}

// GetFilesByUserID извлекает все информации офайлах, связанные с указанным идентификатором пользователя, включая их комментарии. Возвращает список или ошибку.
func (pr *FileRepository) GetFilesByUserID(userID int64) ([]models.FileWithComment, error) {
	var files []models.FileWithComment
	err := pr.db.SelectContext(pr.ctx, &files, getFileByUserIDSQL, models.TypeFile, userID)
	return files, err
}

// GetFileByUserIDAndId извлекает информацию о файле по его идентификатору и идентификатору связанного пользователя. Возвращает текст или ошибку.
func (pr *FileRepository) GetFileByUserIDAndId(userID int64, id int64) (*models.FileContent, error) {
	row := pr.db.QueryRowxContext(pr.ctx, getFileByUserIDAndIDSQL, id, userID)
	err := row.Err()
	if err != nil {
		return nil, err
	}
	var file models.FileContent
	if err = row.StructScan(&file); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.Join(ErrNotExist, err)
		}
		return nil, err
	}
	return &file, err
}

// DeleteFileByUserIDAndID удаляет информацию о файле и связанные с ним комментарии для данного идентификатора пользователя и идентификатора текста.
// Он выполняет два запроса на удаление в рамках транзакции базы данных.
// Возвращает ошибку, если транзакция или запросы завершаются неудачно.
func (pr *FileRepository) DeleteFileByUserIDAndID(userID int64, id int64) error {
	tx, err := pr.db.BeginTxx(pr.ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(pr.ctx, deleteFileByUserIDAndIDSQL, id, userID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(pr.ctx, deleteCommentByContentID, models.TypeFile, id)
	if err != nil {
		return err
	}
	return tx.Commit()
}
