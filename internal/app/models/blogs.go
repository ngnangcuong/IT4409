package models

import "time"

type Blog struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Picture     string    `json:"picture"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	UserID      string    `json:"user_id"`
	TimeCreated time.Time `json:"time_created"`
	LastUpdated time.Time `json:"last_updated"`
}

// Params for Repositories
type GetBlogsParams struct {
	From     int32
	Size     int32
	Sort     string
	Category string
}

type UpdateBlogParams struct {
	ID          string
	Title       string
	Content     string
	Category    string
	LastUpdated time.Time
}

type CreateBlogParams struct {
	ID          string
	Title       string
	Content     string
	Category    string
	UserID      string
	TimeCreated time.Time
	LastUpdated time.Time
}

// Resquest for Services
type GetBlogsRequest struct {
	From     int32  `form:"from" binding:"min=0"`
	Size     int32  `form:"size" binding:"min=0,max=10"`
	Sort     string `form:"sort"`
	Category string `form:"category" binding:"oneof=art science technology cinema desgin food all"`
}

type CreateBlogRequest struct {
	Title    string `json:"title" binding:"required,min=1"`
	Content  string `json:"content" binding:"required,min=1"`
	Category string `json:"category" binding:"required,oneof=art science technology cinema desgin food"`
	Picture  string `json:"picture"`
}

type UpdateBlogRequest struct {
	Title    string `json:"title" binding:"required,min=1"`
	Content  string `json:"content" binding:"required,min=1"`
	Category string `json:"category" binding:"oneof=art science technology cinema desgin food"`
	Picture  string `json:"picture"`
}

// Response for Services
type GetBlogResponse struct {
	Blog     BlogResponse       `json:"blog"`
	Comments []*CommentResponse `json:"comments"`
}

type BlogResponse struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Picture     string       `json:"picture"`
	Content     string       `json:"content"`
	Category    string       `json:"category"`
	User        UserResponse `json:"user"`
	TimeCreated time.Time    `json:"time_created"`
	LastUpdated time.Time    `json:"last_updated"`
}
