package requests

type RegisterUserRequest struct {
	FirstName string `json:"firstName" binding:"required,gte=1,lte=64"`
	LastName  string `json:"lastName" binding:"required,gte=1,lte=64"`
	Email     string `json:"email" binding:"required,email,gte=1,lte=128"`
	Password  string `json:"password" binding:"required,gte=8,lte=32"`
}

type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email,gte=1,lte=128"`
	Password string `json:"password" binding:"required,gte=8,lte=32"`
}

type ChangeUserPasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required,gte=8,lte=32"`
	NewPassword string `json:"newPassword" binding:"required,gte=8,lte=32"`
}
