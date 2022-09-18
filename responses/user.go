package responses

import (
	"dcfs/db/dbo"
	"github.com/google/uuid"
)

type UserDataResponse struct {
	UUID      uuid.UUID `json:"uuid"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"emailName"`
}

type RegisterUserSuccessResponse struct {
	Success bool             `json:"success"`
	Data    UserDataResponse `json:"data"`
}

func NewRegisterUserSuccessResponse(userData dbo.User) *RegisterUserSuccessResponse {
	var r *RegisterUserSuccessResponse = new(RegisterUserSuccessResponse)
	r.Success = true
	r.Data.UUID = userData.UUID
	r.Data.FirstName = userData.FirstName
	r.Data.LastName = userData.LastName
	r.Data.Email = userData.Email
	return r
}
