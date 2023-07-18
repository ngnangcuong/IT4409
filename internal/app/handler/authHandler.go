package handler

import (
	"IT4409/internal/app/models"
	"IT4409/internal/app/services"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

type AuthHandler struct {
	userService  *services.UserServive
	tokenService *services.TokenService
}

func NewAuthHandler(userService *services.UserServive, tokenService *services.TokenService) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		tokenService: tokenService,
	}
}

func (a *AuthHandler) OauthGoogle(c *gin.Context) {
	code := c.Query("code")
	// var pathUrl string = "/"

	// if c.Query("state") != "" {
	// 	pathUrl = c.Query("state")
	// }

	if code == "" {
		errorMessage := models.ErrorResponse{
			Status:       http.StatusUnauthorized,
			ErrorMessage: models.ErrNoPermission.Error(),
		}
		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	tokenResponse, err := a.getGoogleOauthToken(code)
	if err != nil {
		errorMessage := models.ErrorResponse{
			Status:       http.StatusBadGateway,
			ErrorMessage: models.ErrInvalidParameter.Error(),
		}
		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	googleUser, gErr := a.getGoogleUser(tokenResponse.AccessToken, tokenResponse.TokenID)
	if gErr != nil {
		errorMessage := models.ErrorResponse{
			Status:       http.StatusBadGateway,
			ErrorMessage: models.ErrInvalidParameter.Error(),
		}
		c.JSON(errorMessage.Status, errorMessage)
		return
	}

	createUserRequest := models.CreateUserRequest{
		ID:       googleUser.ID,
		Name:     googleUser.GivenName,
		Email:    googleUser.Email,
		Role:     "user",
		Provider: "google",
	}

	successResponse, errorResponse := a.userService.CreateUser(c, createUserRequest)
	if errorResponse != nil {
		c.JSON(errorResponse.Status, errorResponse)
		fmt.Println("Here1")

		return
	}

	user := successResponse.Result.(models.User)
	tokenDetails, tErr := a.tokenService.CreateToken(user.ID)
	if tErr != nil {
		errorResponse := models.ErrorResponse{
			Status:       http.StatusInternalServerError,
			ErrorMessage: models.ErrInternalServerError.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		fmt.Println("Here2")
		return
	}

	successResponse.Result = tokenDetails
	// c.JSON(successResponse.Status, successResponse)
	// c.SetCookie("token", tokenDetails.AccessToken, viper.GetInt("app.at_expires"), "/", "localhost", false, true)
	c.Redirect(http.StatusTemporaryRedirect, fmt.Sprintf("%s?access_token=%s&refresh_token=%s&access_uuid=%s&refresh_uuid=%s",
		viper.GetString("app.frontend_origin"), tokenDetails.AccessToken, tokenDetails.RefreshToken, tokenDetails.AccessUuid, tokenDetails.RefreshUuid))
	return
}

func (a *AuthHandler) Refresh(c *gin.Context) {
	var refreshRequest models.RefreshRequest
	if err := c.ShouldBindJSON(&refreshRequest); err != nil {
		errorResponse := models.ErrorResponse{
			Status:       http.StatusUnauthorized,
			ErrorMessage: models.ErrNoPermission.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	tokenDetails, err := a.tokenService.Refresh(refreshRequest.RefreshToken)
	if err != nil {
		errorResponse := models.ErrorResponse{
			Status:       http.StatusBadRequest,
			ErrorMessage: models.ErrInvalidParameter.Error(),
		}
		c.JSON(errorResponse.Status, errorResponse)
		return
	}

	successResponse := models.SuccessResponse{
		Status: http.StatusCreated,
		Result: tokenDetails,
	}

	c.JSON(successResponse.Status, successResponse)
}

func (a *AuthHandler) Logout(c *gin.Context) {

}

func (a *AuthHandler) getGoogleOauthToken(code string) (*models.GoogleOauthToken, error) {
	const rootURl = "https://oauth2.googleapis.com/token"
	viper.SetConfigFile("./config.yaml")
	err := viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Not found config file")
		} else {
			panic(err.Error())
		}
	}
	values := url.Values{}
	values.Add("grant_type", "authorization_code")
	values.Add("code", code)
	values.Add("client_id", viper.GetString("oauth.google_oauth_client_id"))
	values.Add("client_secret", viper.GetString("oauth.google_oauth_client_secret"))
	values.Add("redirect_uri", viper.GetString("oauth.google_oauth_redirect_url"))

	query := values.Encode()

	req, err := http.NewRequest("POST", rootURl, bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client{
		Timeout: time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("could not retrieve token")
	}

	var resBody bytes.Buffer
	_, err = io.Copy(&resBody, res.Body)
	if err != nil {
		return nil, err
	}

	var GoogleOauthTokenRes map[string]interface{}

	if err := json.Unmarshal(resBody.Bytes(), &GoogleOauthTokenRes); err != nil {
		return nil, err
	}

	tokenBody := &models.GoogleOauthToken{
		AccessToken: GoogleOauthTokenRes["access_token"].(string),
		TokenID:     GoogleOauthTokenRes["id_token"].(string),
	}

	return tokenBody, nil
}

func (a *AuthHandler) getGoogleUser(access_token string, id_token string) (*models.GoogleUserResult, error) {
	rootUrl := fmt.Sprintf("https://www.googleapis.com/oauth2/v1/userinfo?alt=json&access_token=%s", access_token)

	req, err := http.NewRequest("GET", rootUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", id_token))

	client := http.Client{
		Timeout: time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("could not retrieve user")
	}

	var resBody bytes.Buffer
	_, err = io.Copy(&resBody, res.Body)
	if err != nil {
		return nil, err
	}

	var GoogleUserRes map[string]interface{}

	if err := json.Unmarshal(resBody.Bytes(), &GoogleUserRes); err != nil {
		return nil, err
	}

	userBody := &models.GoogleUserResult{
		ID:            GoogleUserRes["id"].(string),
		Email:         GoogleUserRes["email"].(string),
		VerifiedEmail: GoogleUserRes["verified_email"].(bool),
		Name:          GoogleUserRes["name"].(string),
		GivenName:     GoogleUserRes["given_name"].(string),
		Picture:       GoogleUserRes["picture"].(string),
		Locale:        GoogleUserRes["locale"].(string),
	}

	return userBody, nil
}
