package middleware

import (
	"time"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/fourth04/initialize/restfulbygin/config"
	"github.com/fourth04/initialize/restfulbygin/model"
	"github.com/gin-gonic/gin"
	jwtv "gopkg.in/dgrijalva/jwt-go.v3"
)

var realm = config.Cfg.JWTRealm
var key = config.Cfg.JWTKey
var timeout, _ = time.ParseDuration(config.Cfg.JWTTimeout)
var maxRefresh, _ = time.ParseDuration(config.Cfg.JWTMaxRefresh)

var AuthUserMiddleware = &jwt.GinJWTMiddleware{
	Realm:      realm,
	Key:        []byte(key),
	Timeout:    timeout,
	MaxRefresh: maxRefresh,
	Authenticator: func(username string, password string, c *gin.Context) (interface{}, bool) {
		var user model.User
		model.DB.Where("username = ?", username).First(&user)
		if user.ID != 0 {
			return user.Desentitize(), true
		}

		return nil, false
	},
	PayloadFunc: func(data interface{}) jwt.MapClaims {
		rv := jwt.MapClaims{}
		value := data.(map[string]interface{})
		for k, v := range value {
			rv[k] = v
		}
		return rv
	},
	Unauthorized: func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	},
	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup: "header:Authorization",
	// TokenLookup: "query:token",
	// TokenLookup: "cookie:token",

	// TokenHeadName is a string in the header. Default value is "Bearer"
	TokenHeadName: "Bearer",

	// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
	TimeFunc: time.Now,
}

var AuthAdminMiddleware *jwt.GinJWTMiddleware

func init() {
	tmp := jwt.GinJWTMiddleware(*AuthUserMiddleware)
	AuthAdminMiddleware = &tmp
	AuthAdminMiddleware.IdentityHandler = func(claims jwtv.MapClaims) interface{} {
		return claims
	}
	AuthAdminMiddleware.Authorizator = func(id interface{}, c *gin.Context) bool {
		claims := id.(jwtv.MapClaims)
		roleName := claims["RoleName"].(string)
		if roleName == "admin" {
			return true
		}
		return false
	}
}

/*
var AuthAdminMiddleware = &jwt.GinJWTMiddleware{
	Realm:      realm,
	Key:        []byte(key),
	Timeout:    timeout,
	MaxRefresh: maxRefresh,
	Authenticator: func(username string, password string, c *gin.Context) (interface{}, bool) {
		var user model.User
		model.DB.Where("username = ?", username).First(&user)
		if user.ID != 0 {
			return user.Desentitize(), true
		}

		return nil, false
	},
	PayloadFunc: func(data interface{}) jwt.MapClaims {
		rv := jwt.MapClaims{}
		value := data.(map[string]interface{})
		for k, v := range value {
			rv[k] = v
		}
		return rv
	},
	IdentityHandler: func(claims jwtv.MapClaims) interface{} {
		return claims
	},
	Authorizator: func(id interface{}, c *gin.Context) bool {
		claims := id.(jwtv.MapClaims)
		roleName := claims["RoleName"].(string)
		if roleName == "admin" {
			return true
		}
		return false
	},
	Unauthorized: func(c *gin.Context, code int, message string) {
		c.JSON(code, gin.H{
			"code":    code,
			"message": message,
		})
	},
	// TokenLookup is a string in the form of "<source>:<name>" that is used
	// to extract token from the request.
	// Optional. Default value "header:Authorization".
	// Possible values:
	// - "header:<name>"
	// - "query:<name>"
	// - "cookie:<name>"
	TokenLookup: "header:Authorization",
	// TokenLookup: "query:token",
	// TokenLookup: "cookie:token",

	// TokenHeadName is a string in the header. Default value is "Bearer"
	TokenHeadName: "Bearer",

	// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
	TimeFunc: time.Now,
} */
