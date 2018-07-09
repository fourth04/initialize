package utils

import (
	jwt "github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
)

func RateKeyCenByUser(ctx *gin.Context) (string, error) {
	claims := jwt.ExtractClaims(ctx)
	username := claims["Username"].(string)
	return username, nil
}

func RateFormattedGenByUser(ctx *gin.Context) (string, error) {
	claims := jwt.ExtractClaims(ctx)
	rateFormatted := claims["RateFormatted"].(string)
	return rateFormatted, nil
}
