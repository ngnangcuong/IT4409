package services

import (
	"IT4409/internal/app/models"
	repositories "IT4409/internal/app/repositories/user"
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"
)

type UserServive struct {
	userRepo *repositories.UserRepo
	db       *sql.DB
}

func NewUserService(userRepo *repositories.UserRepo, db *sql.DB) *UserServive {
	return &UserServive{
		userRepo: userRepo,
		db:       db,
	}
}

func (u *UserServive) execTx(ctx context.Context, fn func(*repositories.UserRepo) error) error {
	tx, err := u.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	userRepoWithTx := u.userRepo.WithTx(tx)
	err = fn(userRepoWithTx)
	if err != nil {
		if rErr := tx.Rollback(); rErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rErr)
		}
		return err
	}

	return tx.Commit()
}

func (u *UserServive) GetUser(ctx context.Context, id string) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	user, err := u.userRepo.GetUser(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			errorResponse.Status = http.StatusNotFound
			errorResponse.ErrorMessage = models.ErrNoUser.Error()
			return nil, &errorResponse
		}
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Result = user
	successResponse.Status = http.StatusOK
	return &successResponse, nil
}

func (u *UserServive) CreateUser(ctx context.Context, createUserRequest models.CreateUserRequest) (*models.SuccessResponse, *models.ErrorResponse) {
	var successResponse models.SuccessResponse
	var errorResponse models.ErrorResponse

	user, err := u.userRepo.GetUser(ctx, createUserRequest.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			createUserParams := models.CreateUserParams{
				ID:          createUserRequest.ID,
				Name:        createUserRequest.Name,
				Email:       createUserRequest.Email,
				Role:        createUserRequest.Role,
				Provider:    createUserRequest.Provider,
				TimeCreated: time.Now(),
				LastUpdated: time.Now(),
			}
			newUser, createErr := u.userRepo.CreateUser(ctx, createUserParams)
			if createErr != nil {
				errorResponse.Status = http.StatusInternalServerError
				errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
				return nil, &errorResponse
			}

			successResponse.Result = newUser
			successResponse.Status = http.StatusCreated
			return &successResponse, nil
		}
		errorResponse.Status = http.StatusInternalServerError
		errorResponse.ErrorMessage = models.ErrInternalServerError.Error()
		return nil, &errorResponse
	}

	successResponse.Result = user
	successResponse.Status = http.StatusCreated
	return &successResponse, nil
}
