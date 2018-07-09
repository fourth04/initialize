package main

import (
	"log"

	jwt "github.com/appleboy/gin-jwt"
	"github.com/fourth04/initialize/restfulbygin/controller"
	"github.com/fourth04/initialize/restfulbygin/middleware"
	"github.com/fourth04/initialize/restfulbygin/model"
	"github.com/fourth04/initialize/restfulbygin/utils"
	"github.com/gin-gonic/gin"
)

func Cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Next()
	}
}

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	c.JSON(200, gin.H{
		"userID": claims["id"],
		"text":   "Hello World.",
	})
}

func main() {

	defer model.DB.Close()
	r := gin.Default()

	r.Use(Cors())

	rateLimiterMiddleware, err := middleware.NewRateLimiterFromFormatted("1000-M", utils.RateKeyCenByUser, utils.RateFormattedGenByUser)
	if err != nil {
		log.Fatalln("初始化计时器遇错", err)
	}

	r.POST("/login", middleware.AuthUserMiddleware.LoginHandler)
	auth := r.Group("/auth")
	auth.Use(middleware.AuthAdminMiddleware.MiddlewareFunc())
	{
		auth.GET("/hello", rateLimiterMiddleware.Middleware(), helloHandler)
		auth.GET("/refresh_token", middleware.AuthUserMiddleware.RefreshHandler)
	}

	user := r.Group("api/")
	user.Use(middleware.AuthAdminMiddleware.MiddlewareFunc())
	{
		user.POST("/users", controller.PostUser)
		user.GET("/users", controller.GetUsers)
		user.GET("/users/:id", controller.GetUser)
		user.PUT("/users/:id", controller.UpdateUser)
		user.DELETE("/users/:id", controller.DeleteUser)
	}

	sysinfoApi := r.Group("sysinfo/")
	// sysinfoApi.Use(middleware.AuthUserMiddleware.MiddlewareFunc())
	{
		sysinfoApi.GET("if_info_manage", controller.GetIfInfoManage)
		sysinfoApi.GET("if_info_service", controller.GetIfInfoService)
	}

	r.Run(":8080")
}

func OptionsUser(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Methods", "DELETE,POST, PUT")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.Next()
}
