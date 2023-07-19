package handler

import (
	"IT4409/internal/app/models"
	"IT4409/internal/app/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type CommnetHandler struct {
	commentService *services.CommentService
}

func NewCommentHandler(commentService *services.CommentService) *CommnetHandler {
	return &CommnetHandler{
		commentService: commentService,
	}
}

func (c *CommnetHandler) GetComment(ctx *gin.Context) {
	id := ctx.Param("id")
	successResposnse, errorResponse := c.commentService.GetComment(ctx, id)
	if errorResponse != nil {
		ctx.JSON(errorResponse.Status, errorResponse)
		return
	}

	ctx.JSON(successResposnse.Status, successResposnse)
}

func (c *CommnetHandler) GetComments(ctx *gin.Context) {
	blogID := ctx.Query("blog_id")
	successResposnse, errorResponse := c.commentService.GetComments(ctx, blogID)
	if errorResponse != nil {
		ctx.JSON(errorResponse.Status, errorResponse)
		return
	}

	ctx.JSON(successResposnse.Status, successResposnse)
}

func (c *CommnetHandler) UpdateComment(ctx *gin.Context) {
	var updateCommentRequest models.UpdateCommentRequest
	if err := ctx.ShouldBindJSON(&updateCommentRequest); err != nil {
		errorResponse := models.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: models.ErrInvalidParameter.Error(),
		}
		ctx.JSON(errorResponse.Status, errorResponse)
		return
	}

	id := ctx.Param("id")
	userID := ctx.GetString("user_id")
	successResposnse, errorResponse := c.commentService.UpdateComment(ctx, updateCommentRequest, id, userID)
	if errorResponse != nil {
		ctx.JSON(errorResponse.Status, errorResponse)
		return
	}

	ctx.JSON(successResposnse.Status, successResposnse)
}

func (c *CommnetHandler) DeleteComment(ctx *gin.Context) {
	id := ctx.Param("id")
	userID := ctx.GetString("user_id")
	successResposnse, errorResponse := c.commentService.DeleteComment(ctx, id, userID)
	if errorResponse != nil {
		ctx.JSON(errorResponse.Status, errorResponse)
		return
	}

	ctx.JSON(successResposnse.Status, successResposnse)
}

func (c *CommnetHandler) CreateComment(ctx *gin.Context) {
	var createCommentRequest models.CreateCommentRequest
	if err := ctx.ShouldBindJSON(&createCommentRequest); err != nil {
		errorResponse := models.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: models.ErrInvalidParameter.Error(),
		}
		ctx.JSON(errorResponse.Status, errorResponse)
		return
	}
	userID := ctx.GetString("user_id")
	successResposnse, errorResponse := c.commentService.CreateComment(ctx, createCommentRequest, userID)
	if errorResponse != nil {
		ctx.JSON(errorResponse.Status, errorResponse)
		return
	}

	ctx.JSON(successResposnse.Status, successResposnse)
}
