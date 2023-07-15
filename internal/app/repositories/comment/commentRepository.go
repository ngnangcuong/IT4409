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

type CommentRepo struct {
	db DBTX
}

func NewCommentRepo(db DBTX) *CommentRepo {
	return &CommentRepo{
		db: db,
	}
}

func (c *CommentRepo) WithTx(tx *sql.Tx) *CommentRepo {
	return &CommentRepo{
		db: tx,
	}
}

func (c *CommentRepo) GetComment(ctx context.Context, id string) (models.Comment, error) {
	query := `SELECT id, blog_id, user_id, parent_id, content, time_created, last_updated FROM comments WHERE id = $1`
	row := c.db.QueryRowContext(ctx, query, id)
	var comment models.Comment
	err := row.Scan(&comment.ID, &comment.BlogID, &comment.UserID, &comment.ParentID, &comment.Content, &comment.TimeCreated, &comment.LastUpdated)
	return comment, err
}

func (c *CommentRepo) GetCommentBelongToBlog(ctx context.Context, blogID string) ([]models.Comment, error) {
	query := `SELECT id, blog_id, user_id, parent_id, content, time_created, last_updated FROM comments WHERE blog_id = $1`
	rows, err := c.db.QueryContext(ctx, query, blogID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.BlogID, &comment.UserID, &comment.ParentID, &comment.Content,
			&comment.TimeCreated, &comment.LastUpdated); err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, err
}

func (c *CommentRepo) GetCommentForUpdate(ctx context.Context, id string) (models.Comment, error) {
	query := `SELECT id, blog_id, user_id, parent_id, content, time_created, last_updated FROM comments WHERE id = $1 FOR NO KEY UPDATE`
	row := c.db.QueryRowContext(ctx, query, id)
	var comment models.Comment
	err := row.Scan(&comment.ID, &comment.BlogID, &comment.UserID, &comment.ParentID, &comment.Content, &comment.TimeCreated, &comment.LastUpdated)
	return comment, err
}

func (c *CommentRepo) GetCommentsFromParent(ctx context.Context, parentID string) ([]models.Comment, error) {
	query := `SELECT id, blog_id, user_id, parent_id, content, time_created, last_updated FROM comments WHERE parent_id = $1 FOR UPDATE`
	rows, err := c.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.BlogID, &comment.UserID, &comment.ParentID, &comment.Content,
			&comment.TimeCreated, &comment.LastUpdated); err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (c *CommentRepo) DeleteComment(ctx context.Context, id string) error {
	query := `DELETE FROM comments WHERE id = $1`
	_, err := c.db.ExecContext(ctx, query, id)
	return err
}

func (c *CommentRepo) UpdateComment(ctx context.Context, updateCommentParams models.UpdateCommentParams) (models.Comment, error) {
	query := `UPDATE comments SET content = $2, last_updated = $3 WHERE id = $1 RETURNING id, blog_id, user_id, parent_id, content, time_created, last_updated`
	row := c.db.QueryRowContext(ctx, query, updateCommentParams.ID, updateCommentParams.Content, updateCommentParams.LastUpdated)
	var comment models.Comment
	err := row.Scan(&comment.ID, &comment.BlogID, &comment.UserID, &comment.ParentID, &comment.Content,
		&comment.TimeCreated, &comment.LastUpdated)

	return comment, err
}

func (c *CommentRepo) CreateComment(ctx context.Context, createCommentParams models.CreateCommentParams) (models.Comment, error) {
	query := `INSERT INTO comments (id, blog_id, user_id, parent_id, content, time_created, last_updated) VALUES 
	($1, $2, $3, $4, $5, $6, $7) RETURNING id, blog_id, user_id, parent_id, content, time_created, last_updated`
	row := c.db.QueryRowContext(ctx, query, createCommentParams.ID, createCommentParams.BlogID, createCommentParams.UserID,
		createCommentParams.ParentID, createCommentParams.Content, createCommentParams.TimeCreated, createCommentParams.LastUpdated)

	var comment models.Comment
	err := row.Scan(&comment.ID, &comment.BlogID, &comment.UserID, &comment.ParentID, &comment.Content, &comment.TimeCreated, &comment.LastUpdated)
	return comment, err
}
