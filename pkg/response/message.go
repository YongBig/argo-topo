/**
 * @Author: Administrator
 * @Description:
 * @File: message
 * @Date: 2022/5/6 10:49
 */
package response

import (
	"github.com/kataras/iris/v12"
)

func ErrResponse(c iris.Context, err error) {
	c.StatusCode(500)
	_, _ = c.JSON(iris.Map{
		"code":    500,
		"message": err.Error(),
	})
}

func NotFoundResponse(c iris.Context, err error) {
	c.StatusCode(404)
	_, _ = c.JSON(iris.Map{
		"code":    404,
		"message": err.Error(),
	})
}

func SuccessResponse(c iris.Context, data interface{}) {
	c.StatusCode(200)
	_, _ = c.JSON(iris.Map{
		"code": 1,
		"data": data,
	})
}
