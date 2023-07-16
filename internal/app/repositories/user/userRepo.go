package repositories

import (
	"IT4409/internal/app/models"
	"context"
	"database/sql"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type UserRepo struct {
	db DBTX
}

func NewUserRepo(db DBTX) *UserRepo {
	return &UserRepo{
		db: db,
	}
}

func (u *UserRepo) WithTx(tx *sql.Tx) *UserRepo {
	return &UserRepo{
		db: tx,
	}
}

func (u *UserRepo) GetUser(ctx context.Context, id string) (models.User, error) {
	query := `SELECT id, name, email, role, provider, time_created, last_updated FROM users WHERE id = $1`
	row := u.db.QueryRowContext(ctx, query, id)
	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.Provider, &user.TimeCreated, &user.LastUpdated)
	return user, err
}

func (u *UserRepo) GetUserForUpdate(ctx context.Context, id string) (models.User, error) {
	query := `SELECT id, name, email, role, provider, time_created, last_updated FROM users WHERE id = $1 FOR NO KEY UPDATE`
	row := u.db.QueryRowContext(ctx, query, id)
	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.Provider, &user.TimeCreated, &user.LastUpdated)
	return user, err
}

func (u *UserRepo) CreateUser(ctx context.Context, createUserParams models.CreateUserParams) (models.User, error) {
	query := `INSERT INTO users (id, name, email, role, provider, time_created, last_updated) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING id, name, email, role, provider, time_created, last_updated`
	row := u.db.QueryRowContext(ctx, query, createUserParams.ID, createUserParams.Name, createUserParams.Email, createUserParams.Role,
		createUserParams.Provider, createUserParams.TimeCreated, createUserParams.LastUpdated)
	var user models.User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.Role, &user.Provider, &user.TimeCreated, &user.LastUpdated)
	return user, err
}
