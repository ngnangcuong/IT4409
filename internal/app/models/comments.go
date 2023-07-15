package models

import "time"

type Comment struct {
	ID          string
	BlogID      string
	UserID      string
	ParentID    string
	Content     string
	TimeCreated time.Time
	LastUpdated time.Time
}

type UpdateCommentParams struct {
	ID          string
	Content     string
	LastUpdated time.Time
}

type CreateCommentParams struct {
	ID          string
	BlogID      string
	UserID      string
	ParentID    string
	Content     string
	TimeCreated time.Time
	LastUpdated time.Time
}

type UpdateCommentRequest struct {
	Content string `json:"content" binding:"required,min=1"`
}

type CreateCommentRequest struct {
	BlogID   string `json:"blog_id" binding:"required"`
	ParentID string `json:"parent_id"`
	Content  string `json:"content" binding:"required,min=1"`
}
