/**
 * @author 刘荣飞 yes@noxue.com
 * @date 2018/12/31 16:38
 */
package v1

import (
	"github.com/gin-gonic/gin"
	"noxue/utils"
)

func CheckError(c *gin.Context, e interface{}) {
	switch type1 := e.(type) {
	case utils.ApiError:
		c.JSON(type1.Status, gin.H{"status": type1.Code, "msg": type1.Error()})
	case utils.Error:
		c.JSON(500, gin.H{"status": -1, "msg": type1.Error()})
	case error:
		c.JSON(500, gin.H{"status": -1, "msg": type1.Error()})
	default:
		c.JSON(500, gin.H{"status": -1, "msg": e})
	}
	c.Abort()
}
