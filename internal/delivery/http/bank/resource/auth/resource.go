package auth

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,gte=5,lte=100"`
	FirstName string `json:"first_name" validate:"required,alphaunicode,gte=5,lte=100"`
	LastName  string `json:"last_name" validate:"required,alphaunicode,gte=5,lte=100"`
	Age       uint32 `json:"age" validate:"required,gte=18,lt=150"`
}

type UserResponse struct {
	ID        uint64 `json:"user_id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Balance   uint64 `json:"balance"`
	Age       uint32 `json:"age"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=5,lte=100"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UpdatePasswordRequest struct {
	Email       string `json:"email" validate:"required,email"`
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required"`
}

type DeleteUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,gte=5,lte=100"`
}

type UserRequest struct {
	Email string `json:"email" validate:"required,email"`
}
