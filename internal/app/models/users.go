package models

import "time"

type User struct {
	ID          string
	Name        string
	Email       string
	Role        string
	Provider    string
	TimeCreated time.Time
	LastUpdated time.Time
}

type GoogleUserResult struct {
	ID            string
	Email         string
	VerifiedEmail bool
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
	Locale        string
}

type GoogleOauthToken struct {
	AccessToken string
	TokenID     string
}

type CreateUserParams struct {
	ID          string
	Name        string
	Email       string
	Role        string
	Provider    string
	TimeCreated time.Time
	LastUpdated time.Time
}

type CreateUserRequest struct {
	ID       string
	Name     string
	Email    string
	Role     string
	Provider string
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}
