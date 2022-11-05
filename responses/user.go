package responses

import (
	"dcfs/constants"
	"dcfs/db/dbo"
	"github.com/google/uuid"
)

type UserDataResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
}

type UserAuthDataResponse struct {
	UserDataResponse
	Token string `json:"token"`
}

type InvalidCredentialsResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

type UserDataSuccessResponse struct {
	Success bool             `json:"success"`
	Data    UserDataResponse `json:"data"`
}

type LoginSuccessResponse struct {
	Success bool                 `json:"success"`
	Data    UserAuthDataResponse `json:"data"`
}

func NewUserDataSuccessResponse(userData *dbo.User) *UserDataSuccessResponse {
	var r *UserDataSuccessResponse = new(UserDataSuccessResponse)
	r.Success = true
	r.Data.UUID = userData.UUID
	r.Data.FirstName = userData.FirstName
	r.Data.LastName = userData.LastName
	r.Data.Email = userData.Email
	return r
}

func NewLoginSuccessResponse(userData *dbo.User, token string) *LoginSuccessResponse {
	var r *LoginSuccessResponse = new(LoginSuccessResponse)

	r.Success = true
	r.Data.Token = token
	r.Data.UUID = userData.UUID
	r.Data.FirstName = userData.FirstName
	r.Data.LastName = userData.LastName
	r.Data.Email = userData.Email

	return r
}

func NewInvalidCredentialsResponse() *InvalidCredentialsResponse {
	var r *InvalidCredentialsResponse = new(InvalidCredentialsResponse)

	r.Success = false
	r.Message = "Unauthorized"
	r.Code = constants.AUTH_UNAUTHORIZED
	
	return r
}
