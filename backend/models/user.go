package models

type User struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type LoginRequest struct {
	Identifier string `json:"identifier" binding:"required"` // can be email or username
	Password   string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	FirstName string `json:"firstname" binding:"required"`
	LastName  string `json:"lastname" binding:"required"`
	UserName  string `json:"username" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
}

type LoginResponse struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	UserName  string `json:"username"`
	Email     string `json:"email"`
	SKey1     string `json:"skey1"`
	SKey2     string `json:"skey2"`
}
