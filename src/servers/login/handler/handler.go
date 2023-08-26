package handler

import "github.com/gin-gonic/gin"

const (
	ErrCodeSuccess        = 0
	ErrDescSuccess        = "OK"
	ErrCodeUnmarshalParam = 1
	ErrDescUnmarshalParam = "parse param error"
	ErrCodeDBFailed       = 2
	ErrDescDBFailed       = "query db failed"

	ErrCodeInvalidUsername     = 101
	ErrDescInvalidUsername     = "invalid username"
	ErrCodeUsernameExist       = 102
	ErrDescUsernameExist       = "username has been existed"
	ErrCodeInvalidUserOrPasswd = 103
	ErrDescInvalidUserOrPasswd = "login user name or password error"
	ErrCodeValidGateNotFound   = 104
	ErrDescValidGateNotFound   = "available gate not found"
	ErrCodeInvalidNickname     = 105
	ErrDescInvalidNickname     = "invalid nickname"
	ErrCodeNicknameExist       = 106
	ErrDescNicknameExist       = "nickname has been existed"
)

func RegHttpHandler(engine *gin.Engine) {
	engine.POST("/account/create", handlerCreateAccount)
	engine.POST("/account/login", handleLoginAccount)
	engine.POST("/role/create", handleCreateRole)
}
