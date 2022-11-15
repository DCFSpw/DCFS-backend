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

type UserDataSuccessResponse struct {
	Success bool             `json:"success"`
	Data    UserDataResponse `json:"data"`
}

type LoginSuccessResponse struct {
	Success bool                 `json:"success"`
	Data    UserAuthDataResponse `json:"data"`
}

// NewUserDataSuccessResponse - create user data success response
//
// params:
//   - userData: dbo.User object with user data to return
//
// return type:
//   - response: UserDataSuccessResponse with logged user data
func NewUserDataSuccessResponse(userData *dbo.User) *UserDataSuccessResponse {
	var r *UserDataSuccessResponse = new(UserDataSuccessResponse)

	r.Success = true
	r.Data.UUID = userData.UUID
	r.Data.FirstName = userData.FirstName
	r.Data.LastName = userData.LastName
	r.Data.Email = userData.Email

	return r
}

// NewLoginSuccessResponse - create login success response
//
// params:
//   - userData: dbo.User object with user data to return
//   - token: string with JTW bearer authorization token
//
// return type:
//   - response: LoginSuccessResponse with logged user data and auth token
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

// NewInvalidCredentialsResponse - create invalid credentials response
//
// return type:
//   - response: OperationFailureResponse with authorization error information
func NewInvalidCredentialsResponse() *OperationFailureResponse {
	var r *OperationFailureResponse = new(OperationFailureResponse)

	r.Success = false
	r.Message = "Unauthorized"
	r.Code = constants.AUTH_UNAUTHORIZED

	return r
}
