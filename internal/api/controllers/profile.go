package controllers

import (
	"net/http"

	dto "gitxyz/internal/api/dto/request"
	response "gitxyz/internal/api/dto/response"

	"github.com/gin-gonic/gin"
)

func (c *AuthControllerImpl) Profile(ctx *gin.Context) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	user, err := c.service.GetUserByID(userID.(string))
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "profile fetched",
		"data":    response.ToUserResponse(&user),
	})
}

func (c *AuthControllerImpl) GetUserByUsername(ctx *gin.Context) {
	username := ctx.Param("username")
	if username == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	user, err := c.service.GetUserByUsername(username)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "user fetched",
		"data":    response.ToUserResponse(&user),
	})
}

func (c *AuthControllerImpl) UpdateProfile(ctx *gin.Context) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	var request dto.UpdateProfileRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := c.service.UpdateProfile(userID.(string), request.FullName, request.Bio, request.Location, request.Avatar)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "profile updated",
		"data":    response.ToUserResponse(&user),
	})
}

func (c *AuthControllerImpl) ChangePassword(ctx *gin.Context) {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user context missing"})
		return
	}

	var request dto.ChangePasswordRequest
	if err := ctx.ShouldBindBodyWithJSON(&request); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.ChangePassword(userID.(string), request.OldPassword, request.NewPassword); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "password changed successfully"})
}
