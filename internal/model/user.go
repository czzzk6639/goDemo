package model

import "time"

type User struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	Score     int       `json:"score"`
	WinCount  int       `json:"win_count"`
	LoseCount int       `json:"lose_count"`
	CreatedAt time.Time `json:"created_at"`
}

type UserRegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserLoginResponse struct {
	Token string `json:"token"`
	User  *User  `json:"user"`
}

type UserResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
