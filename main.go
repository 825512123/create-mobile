package main

import (
	"create-mobile/api"
	"create-mobile/global"
	"fmt"
	common_go "github.com/825512123/common-go"
	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	Init()
	r := gin.Default()
	r.Use(Core())
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f)

	err2 := os.Mkdir("./mobile", os.ModeDir)
	if err2 != nil {
		fmt.Println("文件夹重复创建", err2)
	}

	r.GET("/mobile", api.Mobile)
	r.POST("/mobile", api.Mobile)

	r.GET("/version", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"msg": "ok", "data": "1.0.1"})
	})
	s := endless.NewServer(fmt.Sprintf(":%d", global.Port), r)
	err := s.ListenAndServe()
	if err != nil {
		log.Printf("server err: %v", err)
	}
}

func Init() {
	global.Port = 8023
	common_go.InitRedis("localhost:6379", "", 1)
}

//Core 解决跨域问题
func Core() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Content-Type,AccessToken,X-CSRF-Token,Authorization,Token")
		c.Header("Access-Control-Allow-Methods", "POST,GET,OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length,Access-Control-Allow-Origin,Access-Control-Allow-Headers,Content-Type")
		c.Header("Access-Control-Allow-Credentials", "True")
		//放行索引options
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		//处理请求
		c.Next()
	}
}
