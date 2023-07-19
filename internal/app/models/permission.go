package models

type Permission struct {
	ID         string
	UserID     string
	ResourceID string
	Action     string
}

type GetPermissionParams struct {
	UserID     string
	ResourceID string
	Action     string
}

type CreatePermissionParams struct {
	UserID     string
	ResourceID string
	Action     string
}
