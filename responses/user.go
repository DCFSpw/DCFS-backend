package responses

import (
	"dcfs/db/dbo"
	"github.com/google/uuid"
)

type UserDataResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
}

type RegisterUserSuccessResponse struct {
	Success bool             `json:"success"`
	Data    UserDataResponse `json:"data"`
}

type LoginSuccessResponse struct {
	Success bool             `json:"success"`
	Token   string           `json:"token"`
	Data    UserDataResponse `json:"data"`
}

func NewRegisterUserSuccessResponse(userData *dbo.User) *RegisterUserSuccessResponse {
	var r *RegisterUserSuccessResponse = new(RegisterUserSuccessResponse)
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
	r.Token = token
	r.Data.UUID = userData.UUID
	r.Data.FirstName = userData.FirstName
	r.Data.LastName = userData.LastName
	r.Data.Email = userData.Email
	return r
}
