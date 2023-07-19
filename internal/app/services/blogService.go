package services

import (
	"IT4409/internal/app/models"
	blogrepo "IT4409/internal/app/repositories/blog"
	commentrepo "IT4409/internal/app/repositories/comment"
	permissionrepo "IT4409/internal/app/repositories/permission"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/twinj/uuid"
)

type BlogService struct {
	blogRepo       *blogrepo.BlogRepo
	commentRepo    *commentrepo.CommentRepo
	permissionRepo *permissionrepo.PermissionRepo
	db             *sql.DB
}

func NewBlogService(blogRepo *blogrepo.BlogRepo, commentRepo *commentrepo.CommentRepo, permissionRepo *permissionrepo.PermissionRepo, db *sql.DB) *BlogService {
	return &BlogService{
		blogRepo:       blogRepo,
		commentRepo:    commentRepo,
		permissionRepo: permissionRepo,
		db:             db,
	}
}

func (b *BlogService) execTx(ctx context.Context, fn func(*blogrepo.BlogRepo, *permissionrepo.PermissionRepo) error) error {
	tx, err := b.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	blogRepoWithTx := b.blogRepo.WithTx(tx)
	permissionRepoWithTx := b.permissionRepo.WithTx(tx)
	err = fn(blogRepoWithTx, permissionRepoWithTx)
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
		From:     getBlogsRequest.From,
		Size:     getBlogsRequest.Size,
		Sort:     getBlogsRequest.Sort,
		Category: getBlogsRequest.Category,
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

func (b *BlogService) CreateBlog(ctx context.Context, createBlogRequest models.CreateBlogRequest, userID string) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	if createBlogRequest.Content == "" || createBlogRequest.Title == "" {
		errorResponse.Status = http.StatusBadRequest
		errorResponse.ErrorMessage = models.ErrInvalidParameter.Error()
		return nil, &errorResponse
	}

	err := b.execTx(ctx, func(br *blogrepo.BlogRepo, pr *permissionrepo.PermissionRepo) error {
		createBlogParams := models.CreateBlogParams{
			ID:          uuid.NewV4().String(),
			Title:       createBlogRequest.Title,
			Content:     createBlogRequest.Content,
			Category:    createBlogRequest.Category,
			UserID:      userID,
			TimeCreated: time.Now(),
			LastUpdated: time.Now(),
		}

		newBlog, err := br.CreateBlog(ctx, createBlogParams)
		if err != nil {
			return err
		}

		createPermissionParams := models.CreatePermissionParams{
			UserID:     userID,
			ResourceID: newBlog.ID,
			Action:     "Update",
		}

		_, pErr := pr.CreatePermission(ctx, createPermissionParams)
		if pErr != nil {
			return pErr
		}

		successResponse.Result = newBlog
		successResponse.Status = http.StatusCreated
		return nil
	})

	if err != nil {
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}
	return &successResponse, nil
}

func (b *BlogService) UpdateBlog(ctx context.Context, updateBlogRequest models.UpdateBlogRequest, id, userID string) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	if updateBlogRequest.Title == "" || updateBlogRequest.Content == "" {
		errorResponse.Status = http.StatusBadRequest
		errorResponse.ErrorMessage = models.ErrInvalidParameter.Error()
		return nil, &errorResponse
	}

	err := b.execTx(ctx, func(br *blogrepo.BlogRepo, pr *permissionrepo.PermissionRepo) error {
		blog, err := br.GetBlogForUpdate(ctx, id)
		if err != nil {
			return err
		}

		getPermissionParams := models.GetPermissionParams{
			UserID:     userID,
			ResourceID: blog.ID,
			Action:     "Update",
		}
		_, pErr := pr.GetPermission(ctx, getPermissionParams)
		if pErr != nil {
			if pErr == sql.ErrNoRows {
				return models.ErrNoPermission
			}
			return pErr
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

		if err == models.ErrNoPermission {
			errorResponse.Status = http.StatusUnauthorized
			errorResponse.ErrorMessage = err.Error()
			return nil, &errorResponse
		}
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Status = http.StatusOK
	return &successResponse, nil
}
