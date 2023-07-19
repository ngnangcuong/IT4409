package services

import (
	"IT4409/internal/app/models"
	blogrepo "IT4409/internal/app/repositories/blog"
	commentrepo "IT4409/internal/app/repositories/comment"
	permissionrepo "IT4409/internal/app/repositories/permission"
	userrepo "IT4409/internal/app/repositories/user"
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
	userRepo       *userrepo.UserRepo
	permissionRepo *permissionrepo.PermissionRepo
	db             *sql.DB
}

func NewBlogService(blogRepo *blogrepo.BlogRepo, commentRepo *commentrepo.CommentRepo, permissionRepo *permissionrepo.PermissionRepo,
	userRepo *userrepo.UserRepo, db *sql.DB) *BlogService {
	return &BlogService{
		blogRepo:       blogRepo,
		commentRepo:    commentRepo,
		permissionRepo: permissionRepo,
		userRepo:       userRepo,
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

	user, err := b.userRepo.GetUser(ctx, blog.UserID)
	if err != nil {
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		// return nil, &errorResponse
	}
	comments, err := b.commentRepo.GetCommentBelongToBlog(ctx, id)
	if err != nil {
		if err != sql.ErrNoRows {
			errorResponse.Status = http.StatusInternalServerError
			errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
			return nil, &errorResponse
		}
	}
	mapCommentResponse := make(map[string][]*models.CommentResponse)
	var listCommentResponse []*models.CommentResponse
	for _, comment := range comments {
		commentResponse := models.CommentResponse{
			ID:          comment.ID,
			BlogID:      comment.BlogID,
			Content:     comment.Content,
			TimeCreated: comment.TimeCreated,
			LastUpdated: comment.LastUpdated,
		}
		if comment.ID == comment.ParentID {
			mapCommentResponse[comment.ID] = []*models.CommentResponse{}
			listCommentResponse = append(listCommentResponse, &commentResponse)
		} else {
			val, ok := mapCommentResponse[comment.ParentID]
			if ok {
				val = append(val, &commentResponse)
				mapCommentResponse[comment.ParentID] = val
			} else {
				mapCommentResponse[comment.ParentID] = []*models.CommentResponse{&commentResponse}
			}
		}
	}
	for _, comment := range listCommentResponse {
		comment.ChildComments = mapCommentResponse[comment.ID]
	}
	blogResponse := models.BlogResponse{
		ID:          blog.ID,
		Title:       blog.Title,
		Picture:     blog.Picture,
		Content:     blog.Content,
		Category:    blog.Category,
		TimeCreated: blog.TimeCreated,
		LastUpdated: blog.LastUpdated,
		User: models.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Name,
		},
	}
	getBlogResponse := models.GetBlogResponse{
		Blog:     blogResponse,
		Comments: listCommentResponse,
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
	if getBlogsParams.Category == "all" {
		getBlogsParams.Category = ""
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
	var blogsResponse []models.BlogResponse
	for _, blog := range blogs {
		user, err := b.userRepo.GetUser(ctx, blog.UserID)
		if err != nil {
			errorResponse.Status = http.StatusInternalServerError
			errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
			// return nil, &errorResponse
		}
		blogsResponse = append(blogsResponse, models.BlogResponse{
			ID:          blog.ID,
			Title:       blog.Title,
			Picture:     blog.Picture,
			Content:     blog.Content,
			Category:    blog.Category,
			TimeCreated: blog.TimeCreated,
			LastUpdated: blog.LastUpdated,
			User: models.UserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
			},
		})
	}
	successResponse.Result = blogsResponse
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
			Picture:     createBlogRequest.Picture,
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
