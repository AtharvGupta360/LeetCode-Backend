package middleware

import(
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/gupta/leetcode-judge/internal/common"
)

func Recovery() gin.HandlerFunc{
	return func(c * gin.Context){
		defer func(){
			if err := recover(); err!=nil{
				common.Logger.Errorf("panic error %v",err)
				common.Error(c,http.StatusInternalServerError,"internal server error","INTERNAL_ERROR")
				c.Abort()
			}
		}()
		c.Next()
	}
}