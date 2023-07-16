package services

import (
	"IT4409/internal/app/models"
	repositories "IT4409/internal/app/repositories/token"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/twinj/uuid"
)

type TokenService struct {
	tokenRepo     *repositories.TokenRepo
	accessSecret  string
	refreshSecret string
	atExpires     int64
	rtExpires     int64
}

func NewTokenService(tokenRepo *repositories.TokenRepo, atExpires, rtExpires int64, accessSecret, refreshSecret string) *TokenService {
	return &TokenService{
		tokenRepo:     tokenRepo,
		accessSecret:  accessSecret,
		refreshSecret: refreshSecret,
		atExpires:     atExpires,
		rtExpires:     rtExpires,
	}
}

func (t *TokenService) CreateToken(userId string) (*models.TokenDetails, error) {
	td := models.TokenDetails{}
	td.AccessUuid = uuid.NewV4().String()
	td.AtExpires = t.atExpires
	td.RefreshUuid = uuid.NewV4().String()
	td.RtExpires = t.rtExpires

	var err error

	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["user_id"] = userId
	atClaims["access_uuid"] = td.AccessUuid
	atClaims["exp"] = td.AtExpires

	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	td.AccessToken, err = at.SignedString([]byte(t.accessSecret))
	if err != nil {
		return nil, err
	}

	rtClaims := jwt.MapClaims{}
	rtClaims["user_id"] = userId
	rtClaims["refresh_uuid"] = td.AccessUuid
	rtClaims["exp"] = td.RtExpires

	rt := jwt.NewWithClaims(jwt.SigningMethodHS256, rtClaims)
	td.RefreshToken, err = rt.SignedString([]byte(t.refreshSecret))
	if err != nil {
		return nil, err
	}

	if err := t.tokenRepo.StoreToken(userId, td.AccessUuid, time.Unix(td.AtExpires, 0)); err != nil {
		return nil, err
	}

	if err := t.tokenRepo.StoreToken(userId, td.RefreshUuid, time.Unix(td.RtExpires, 0)); err != nil {
		return nil, err
	}

	return &td, nil
}

func (t *TokenService) Refresh(refreshToken string) (*models.TokenDetails, error) {
	token, err := jwt.Parse(refreshToken, func(_token *jwt.Token) (interface{}, error) {
		if _, ok := _token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", _token.Header["alg"])
		}

		return []byte(t.refreshSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return nil, fmt.Errorf("Token has expired")
	}

	rtClaims := token.Claims.(jwt.MapClaims)

	userID := rtClaims["user_id"].(string)

	refreshUuid, ok := rtClaims["refresh_uuid"].(string)
	if !ok {
		return nil, fmt.Errorf("Token has expired")
	}

	deleted, deleteErr := t.tokenRepo.DeleteToken(refreshUuid)
	if deleted == 0 || deleteErr != nil {
		return nil, fmt.Errorf("Delete token error")
	}

	return t.CreateToken(userID)
}

func (t *TokenService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(_token *jwt.Token) (interface{}, error) {
		if _, ok := _token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", _token.Header["alg"])
		}

		return []byte(t.accessSecret), nil
	})
	if err != nil {
		return nil, err
	}

	if _, ok := token.Claims.(jwt.MapClaims); !ok || !token.Valid {
		return nil, fmt.Errorf("Token has expired")
	}

	return token, nil
}

func (t *TokenService) DeleteToken(tokenUuid string) (int64, error) {
	return t.tokenRepo.DeleteToken(tokenUuid)
}

func (t *TokenService) ExtractTokenFromRequest(r *http.Request) string {
	bearToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}

	return ""
}
