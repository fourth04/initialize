package controller

import (
	"net/http"

	"github.com/fourth04/initialize/sysinfo"
	"github.com/gin-gonic/gin"
)

func GetIfInfoManage(c *gin.Context) {
	adapters := sysinfo.GetIfInfoManage("192.168.128.114")
	c.JSON(http.StatusOK, gin.H{"success": adapters})
}

func GetIfInfoService(c *gin.Context) {
	adapters := sysinfo.GetIfInfoService("192.168.128.114")
	c.JSON(http.StatusOK, gin.H{"success": adapters})
}
