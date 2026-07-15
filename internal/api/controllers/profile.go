package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *AuthControllerImpl) Profile(ctx *gin.Context) {
	userID, _ := ctx.Get("user_id")
	username, _ := ctx.Get("username")
	email, _ := ctx.Get("email")

	ctx.JSON(http.StatusOK, gin.H{
		"message": "profile fetched",
		"data": gin.H{
			"user_id": userID,
			"username": username,
			"email": email,
		},
	})
}
