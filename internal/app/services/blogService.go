package services

import (
	"IT4409/internal/app/models"
	blogrepo "IT4409/internal/app/repositories/blog"
	commentrepo "IT4409/internal/app/repositories/comment"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/twinj/uuid"
)

type BlogService struct {
	blogRepo    *blogrepo.BlogRepo
	commentRepo *commentrepo.CommentRepo
	db          *sql.DB
}

func NewBlogService(blogRepo *blogrepo.BlogRepo, commentRepo *commentrepo.CommentRepo, db *sql.DB) *BlogService {
	return &BlogService{
		blogRepo:    blogRepo,
		commentRepo: commentRepo,
		db:          db,
	}
}

func (b *BlogService) execTx(ctx context.Context, fn func(*blogrepo.BlogRepo) error) error {
	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	blogRepoWithTx := b.blogRepo.WithTx(tx)
	err = fn(blogRepoWithTx)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rErr)
		}
		return err
	}

	return tx.Commit()
}

func (b *BlogService) GetBlog(ctx context.Context, id string) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	blog, err := b.blogRepo.GetBlog(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = models.ErrNoBlog.Error()
			return nil, &errorResponse
		}

		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	comments, err := b.commentRepo.GetCommentBelongToBlog(ctx, id)
	if err != nil {
		if err != sql.ErrNoRows {
			errorResponse.Status = http.StatusInternalServerError
			errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
			return nil, &errorResponse
		}
	}

	getBlogResponse := models.GetBlogResponse{
		Blog:     blog,
		Comments: comments,
	}
	successResponse.Result = getBlogResponse
	successResponse.Status = http.StatusOK
	return &successResponse, nil
}

func (b *BlogService) GetBlogs(ctx context.Context, getBlogsRequest models.GetBlogsRequest) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	getBlogsParams := models.GetBlogsParams{
		From: getBlogsRequest.From,
		Size: getBlogsRequest.Size,
		Sort: getBlogsRequest.Sort,
	}

	blogs, err := b.blogRepo.GetBlogs(ctx, getBlogsParams)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = models.ErrNoBlog.Error()
			return nil, &errorResponse
		}
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Result = blogs
	successResponse.Status = http.StatusOK
	return &successResponse, nil
}

func (b *BlogService) CreateBlog(ctx context.Context, createBlogRequest models.CreateBlogRequest) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	if createBlogRequest.Content == "" || createBlogRequest.Title == "" {
		errorResponse.Status = http.StatusBadRequest
		errorResponse.ErrorMessage = models.ErrInvalidParameter.Error()
		return nil, &errorResponse
	}

	createBlogParams := models.CreateBlogParams{
		ID:          uuid.NewV4().String(),
		Title:       createBlogRequest.Title,
		Content:     createBlogRequest.Content,
		Category:    createBlogRequest.Category,
		UserID:      "",
		TimeCreated: time.Now(),
		LastUpdated: time.Now(),
	}

	newBlog, err := b.blogRepo.CreateBlog(ctx, createBlogParams)
	if err != nil {
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Result = newBlog
	successResponse.Status = http.StatusCreated
	return &successResponse, nil
}

func (b *BlogService) UpdateBlog(ctx context.Context, updateBlogRequest models.UpdateBlogRequest, id string) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	if updateBlogRequest.Title == "" || updateBlogRequest.Content == "" {
		errorResponse.Status = http.StatusBadRequest
		errorResponse.ErrorMessage = models.ErrInvalidParameter.Error()
		return nil, &errorResponse
	}

	err := b.execTx(ctx, func(br *blogrepo.BlogRepo) error {
		blog, err := br.GetBlogForUpdate(ctx, id)
		if err != nil {
			return err
		}

		updateBlogParams := models.UpdateBlogParams{
			ID:          blog.ID,
			Title:       updateBlogRequest.Title,
			Content:     updateBlogRequest.Content,
			Category:    updateBlogRequest.Category,
			LastUpdated: time.Now(),
		}

		_, uErr := br.UpdateBlog(ctx, updateBlogParams)
		return uErr
	})

	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = models.ErrNoBlog.Error()
			return nil, &errorResponse
		}

		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Status = http.StatusOK
	return &successResponse, nil
}
