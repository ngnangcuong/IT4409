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

type PermissionRepo struct {
	db DBTX
}

func NewPermissionRepo(db DBTX) *PermissionRepo {
	return &PermissionRepo{
		db: db,
	}
}

func (p *PermissionRepo) WithTx(tx *sql.Tx) *PermissionRepo {
	return &PermissionRepo{
		db: tx,
	}
}

func (p *PermissionRepo) GetPermission(ctx context.Context, getPermissionParams models.GetPermissionParams) (models.Permission, error) {
	query := `SELECT id, user_id, resource_id, action FROM permissions WHERE user_id = $1 AND resource_id = $2 AND action = $3`
	row := p.db.QueryRowContext(ctx, query, getPermissionParams.UserID, getPermissionParams.ResourceID, getPermissionParams.Action)
	var permission models.Permission
	err := row.Scan(&permission.ID, &permission.UserID, &permission.ResourceID, &permission.Action)
	return permission, err
}

func (p *PermissionRepo) GetPermissionByResourceID(ctx context.Context, resourceID string) (models.Permission, error) {
	query := `SELECT id, user_id, resource_id, action FROM permissions WHERE resource_id = $1`
	row := p.db.QueryRowContext(ctx, query, resourceID)
	var permission models.Permission
	err := row.Scan(&permission.ID, &permission.UserID, &permission.ResourceID, &permission.Action)
	return permission, err
}

func (p *PermissionRepo) CreatePermission(ctx context.Context, createPermissionParams models.CreatePermissionParams) (models.Permission, error) {
	query := `INSERT INTO permissions (user_id, resource_id, action) VALUES ($1, $2, $3) RETURNING id, user_id, resource_id, action`
	row := p.db.QueryRowContext(ctx, query, createPermissionParams.UserID, createPermissionParams.ResourceID, createPermissionParams.Action)
	var permission models.Permission
	err := row.Scan(&permission.ID, &permission.UserID, &permission.ResourceID, &permission.Action)
	return permission, err
}

func (p *PermissionRepo) DeletePermission(ctx context.Context, resourceID string) error {
	query := `DELETE FROM permissions WHERE resource_id = $1`
	_, err := p.db.ExecContext(ctx, query, resourceID)
	return err
}
