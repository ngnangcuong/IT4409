package services

import (
	"IT4409/internal/app/models"
	repositories "IT4409/internal/app/repositories/comment"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/lib/pq"
	"github.com/twinj/uuid"
)

type CommentService struct {
	commentRepo *repositories.CommentRepo
	db          *sql.DB
}

func NewCommentService(commentRepo *repositories.CommentRepo, db *sql.DB) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		db:          db,
	}
}

func (c *CommentService) execTx(ctx context.Context, fn func(*repositories.CommentRepo) error) error {
	tx, err := c.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	commentRepoWithTx := c.commentRepo.WithTx(tx)
	err = fn(commentRepoWithTx)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rErr)
		}
		return err
	}

	return tx.Commit()
}

func (c *CommentService) GetComment(ctx context.Context, id string) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	comment, err := c.commentRepo.GetComment(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = models.ErrNoComment.Error()
			return nil, &errorResponse
		}

		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Result = comment
	successResponse.Status = http.StatusOK
	return &successResponse, nil
}

func (c *CommentService) GetComments(ctx context.Context, blogID string) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	commets, err := c.commentRepo.GetCommentBelongToBlog(ctx, blogID)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = models.ErrNoComment.Error()
			return nil, &errorResponse
		}

		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Result = commets
	successResponse.Status = http.StatusOK
	return &successResponse, nil
}

func (c *CommentService) UpdateComment(ctx context.Context, updateCommentRequest models.UpdateCommentRequest, id string) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	if updateCommentRequest.Content == "" {
		errorResponse.Status = http.StatusBadRequest
		errorResponse.ErrorMessage = models.ErrInvalidParameter.Error()
		return nil, &errorResponse
	}

	err := c.execTx(ctx, func(cr *repositories.CommentRepo) error {
		comment, err := cr.GetCommentForUpdate(ctx, id)
		if err != nil {
			return err
		}

		updateCommentParams := models.UpdateCommentParams{
			ID:          comment.ID,
			Content:     updateCommentRequest.Content,
			LastUpdated: time.Now(),
		}

		_, uErr := cr.UpdateComment(ctx, updateCommentParams)
		return uErr
	})

	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
			return nil, &errorResponse
		}
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Status = http.StatusOK
	return &successResponse, nil
}

func (c *CommentService) CreateComment(ctx context.Context, createCommentRequest models.CreateCommentRequest) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	if createCommentRequest.Content == "" {
		errorResponse.Status = http.StatusBadRequest
		errorResponse.ErrorMessage = models.ErrInvalidParameter.Error()
		return nil, &errorResponse
	}

	createCommentParams := models.CreateCommentParams{
		ID:          uuid.NewV4().String(),
		BlogID:      createCommentRequest.BlogID,
		UserID:      "",
		Content:     createCommentRequest.Content,
		TimeCreated: time.Now(),
		LastUpdated: time.Now(),
	}
	if createCommentRequest.ParentID == "" {
		createCommentParams.ParentID = createCommentParams.ID
	} else {
		createCommentParams.ParentID = createCommentRequest.ParentID
	}

	comment, err := c.commentRepo.CreateComment(ctx, createCommentParams)
	if err != nil {
		fmt.Println(err)
		if _, ok := err.(*pq.Error); ok {
			errorResponse.Status = http.StatusBadRequest
			errorResponse.ErrorMessage = models.ErrInvalidParameter.Error()
		}
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Result = comment
	successResponse.Status = http.StatusOK
	return &successResponse, nil
}

func (c *CommentService) DeleteComment(ctx context.Context, id string) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	err := c.execTx(ctx, func(cr *repositories.CommentRepo) error {
		comments, err := cr.GetCommentsFromParent(ctx, id)
		if err != nil {
			return err
		}

		for _, comment := range comments {
			err := cr.DeleteComment(ctx, comment.ID)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = models.ErrNoComment.Error()
			return nil, &errorResponse
		}
		errorResponse.Status = http.StatusBadRequest
		errorResponse.ErrorMessage = models.ErrInvalidParameter.Error()
		return nil, &errorResponse
	}

	successResponse.Status = http.StatusOK
	return &successResponse, nil
}
