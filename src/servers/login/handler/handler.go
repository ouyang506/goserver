package handler

import "github.com/gin-gonic/gin"

const (
	ErrCodeSuccess        = 0
	ErrDescSuccess        = "OK"
	ErrCodeUnmarshalParam = 1
	ErrDescUnmarshalParam = "parse param error"

	ErrCodeInvalidUserOrPasswd = 101
	ErrDescInvalidUserOrPasswd = "login user name or password error"
)

func RegHttpHandler(engine *gin.Engine) {
	engine.POST("/login", handlerAccountLogin)
}
