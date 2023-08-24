package handler

import "github.com/gin-gonic/gin"

const (
	ErrCodeSuccess        = 0
	ErrDescSuccess        = "OK"
	ErrCodeUnmarshalParam = 1
	ErrDescUnmarshalParam = "parse param error"
	ErrCodeDBFailed       = 2
	ErrDescDBFailed       = "query db failed"

	ErrCodeInvalidUserOrPasswd = 101
	ErrDescInvalidUserOrPasswd = "login user name or password error"
	ErrCodeValidGateNotFound   = 102
	ErrDescValidGateNotFound   = "available gate not found"
)

func RegHttpHandler(engine *gin.Engine) {
	engine.POST("/login", handlerAccountLogin)
}
