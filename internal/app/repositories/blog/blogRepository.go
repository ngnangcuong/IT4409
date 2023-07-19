package repositories

import (
	"IT4409/internal/app/models"
	"context"
	"database/sql"
	"fmt"
)

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

type BlogRepo struct {
	db DBTX
}

func NewBlogRepo(db DBTX) *BlogRepo {
	return &BlogRepo{
		db: db,
	}
}

func (b *BlogRepo) WithTx(tx *sql.Tx) *BlogRepo {
	return &BlogRepo{
		db: tx,
	}
}

func (b *BlogRepo) GetBlogs(ctx context.Context, getBlogsParams models.GetBlogsParams) ([]models.Blog, error) {
	query := `SELECT id, user_id, title, content, category, time_created, last_updated FROM blogs WHERE category ~ $3 OFFSET $1 LIMIT $2 `
	rows, err := b.db.QueryContext(ctx, query, getBlogsParams.From, getBlogsParams.Size, getBlogsParams.Category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blogs []models.Blog
	for rows.Next() {
		var blog models.Blog
		if err := rows.Scan(&blog.ID, &blog.UserID, &blog.Title, &blog.Content, &blog.Category, &blog.TimeCreated, &blog.LastUpdated); err != nil {
			return nil, err
		}
		fmt.Println(blog)
		blogs = append(blogs, blog)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return blogs, nil
}

func (b *BlogRepo) GetBlog(ctx context.Context, id string) (models.Blog, error) {
	query := `SELECT id, user_id, title, content, category, time_created, last_updated FROM blogs WHERE id = $1`
	row := b.db.QueryRowContext(ctx, query, id)
	var blog models.Blog
	err := row.Scan(&blog.ID, &blog.UserID, &blog.Title, &blog.Content, &blog.Category, &blog.TimeCreated, &blog.LastUpdated)
	return blog, err
}

func (b *BlogRepo) GetBlogForUpdate(ctx context.Context, id string) (models.Blog, error) {
	query := `SELECT id, user_id, title, content, category, time_created, last_updated FROM blogs WHERE id = $1 FOR NO KEY UPDATE`
	row := b.db.QueryRowContext(ctx, query, id)
	var blog models.Blog
	err := row.Scan(&blog.ID, &blog.UserID, &blog.Title, &blog.Content, &blog.Category, &blog.TimeCreated, &blog.LastUpdated)
	return blog, err
}

func (b *BlogRepo) UpdateBlog(ctx context.Context, updateBlogParams models.UpdateBlogParams) (models.Blog, error) {
	query := `UPDATE blogs SET title = $2, content = $3, last_updated = $4 WHERE id = $1 RETURNING id, user_id, title, content, category, time_created, last_updated`
	row := b.db.QueryRowContext(ctx, query, updateBlogParams.ID, updateBlogParams.Title, updateBlogParams.Content, updateBlogParams.LastUpdated)
	var blog models.Blog
	err := row.Scan(&blog.ID, &blog.UserID, &blog.Title, &blog.Content, &blog.Category, &blog.TimeCreated, &blog.LastUpdated)
	return blog, err
}

func (b *BlogRepo) DeleteBlog(ctx context.Context, id string) error {
	query := `DELETE FROM blogs WHERE id = $1`
	_, err := b.db.ExecContext(ctx, query, id)
	return err
}

func (b *BlogRepo) CreateBlog(ctx context.Context, createBlogParams models.CreateBlogParams) (models.Blog, error) {
	query := `INSERT INTO blogs (id, user_id, title, content, category, time_created, last_updated)
	VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, user_id, title, content, category, time_created, last_updated`
	row := b.db.QueryRowContext(ctx, query, createBlogParams.ID, createBlogParams.UserID, createBlogParams.Title, createBlogParams.Content,
		createBlogParams.Category, createBlogParams.TimeCreated, createBlogParams.LastUpdated)
	var blog models.Blog
	err := row.Scan(&blog.ID, &blog.UserID, &blog.Title, &blog.Content, &blog.Category, &blog.TimeCreated, &blog.LastUpdated)
	return blog, err
}
