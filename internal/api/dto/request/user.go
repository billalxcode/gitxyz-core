package dto

type RegisterRequest struct {
	FullName string `json:"full_name" binding:"required,min=4"`
	Username string `json:"username" binding:"required,min=4"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=24"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	FullName string `json:"full_name" binding:"omitempty,min=4"`
	Bio      string `json:"bio" binding:"omitempty"`
	Location string `json:"location" binding:"omitempty"`
	Avatar   string `json:"avatar" binding:"omitempty"`
}

type UpdateBioRequest struct {
	Bio string `json:"bio" binding:"required"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=24"`
}

type AddSSHKeyRequest struct {
	Title     string `json:"title" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

type CreateTokenRequest struct {
	Name   string `json:"name" binding:"required"`
	Scopes string `json:"scopes"`
}
