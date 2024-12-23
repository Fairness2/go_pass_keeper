package repositories

import (
	"context"
	"database/sql"
	"errors"
	"passkeeper/internal/models"
)

const (
	getUserByIDSQL    = "SELECT id, login, password_hash FROM t_user WHERE id = $1"
	getUserByLoginSQL = "SELECT id, login, password_hash FROM t_user WHERE login = $1"
	createUserSQL     = "INSERT INTO t_user (login, password_hash) VALUES (:login, :password_hash) RETURNING id"
	userExistsSQL     = "SELECT true FROM t_user WHERE login = $1"
)

// UserRepository представляет собой хранилище для управления данными пользователя.
type UserRepository struct {
	// db пул соединений с базой данных, которыми может пользоваться хранилище
	db SQLExecutor
}

// NewUserRepository initializes and returns a new UserRepository with the given context and SQLExecutor.
func NewUserRepository(db SQLExecutor) *UserRepository { // TODO заменить на интерфейс
	return &UserRepository{
		db: db,
	}
}

// UserExists проверяем наличие пользователя
func (r *UserRepository) UserExists(ctx context.Context, login string) error {
	var exists bool
	err := r.db.QueryRowContext(ctx, userExistsSQL, login).Scan(&exists)
	// Если у нас нет записей, то возвращаем false
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return ErrNotExist
	}
	if err != nil {
		return err
	}
	return nil
}

// CreateUser вставляем нового пользователя и присваиваем ему id
func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	smth, err := r.db.PrepareNamed(createUserSQL)
	if err != nil {
		return err
	}
	row := smth.QueryRowxContext(ctx, user)
	return row.Scan(&user.ID)
}

// GetUserByLogin извлекает пользователя на основе его логина из базы данных.
// Возвращает пользователя, логическое значение, если найдено, и ошибку.
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRowxContext(ctx, getUserByLoginSQL, login).StructScan(&user)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Join(ErrNotExist, err)
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID извлекает пользователя по его уникальному идентификатору из базы данных.
// Возвращает пользователя, логическое значение, указывающее на существование, и ошибку.
func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User

	err := r.db.QueryRowxContext(ctx, getUserByIDSQL, id).StructScan(&user)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return nil, errors.Join(ErrNotExist, err)
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}
