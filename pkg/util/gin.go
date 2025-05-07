package util

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/json"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

// gin 上下文 key
const (
	ReqBodyKey = "req_body"
	ResBodyKey = "res_body"
)

// ResponseResult 响应结果
type ResponseResult struct {
	Code    int         `json:"code"`              // 错误码
	Message string      `json:"message,omitempty"` // 错误消息
	Data    interface{} `json:"data,omitempty"`    // 响应数据
}

// ResOK 响应成功
func ResOK(c *gin.Context) {
	ResSuccess(c, nil)
}

// ResSuccess 响应成功
func ResSuccess(c *gin.Context, data interface{}) {
	result := &ResponseResult{
		Code: 200,
		Data: data,
	}

	if v, ok := data.(string); ok && v != "" {
		result.Message = v
		result.Data = nil
	}

	ResJSON(c, http.StatusOK, result)
}

// ResPage 响应分页数据
func ResPage(c *gin.Context, list interface{}, pagination interface{}) {
	result := &ResponseResult{
		Code: 0,
		Data: gin.H{
			"list":       list,
			"pagination": pagination,
		},
	}
	ResJSON(c, http.StatusOK, result)
}

// ResError 响应错误
func ResError(c *gin.Context, err error) {
	ierr := errors.FromError(err)
	httpCode := errors.StatusCode(ierr)

	res := &ResponseResult{
		Code:    ierr.Code(),
		Message: ierr.Message(),
	}

	if httpCode >= 500 {
		ctx := c.Request.Context()
		ctx = logging.NewTag(ctx, logging.TagKeySystem)
		ctx = logging.NewStack(ctx, fmt.Sprintf("%+v", err))
		logging.Context(ctx).Error("服务器内部错误", zap.Error(err))
	}

	ResJSON(c, httpCode, res)
}

func ResJSON(c *gin.Context, status int, v interface{}) {
	buf, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	c.Set(ResBodyKey, buf)
	c.Data(status, "application/json; charset=utf-8", buf)
	c.Abort()
}

// ParseJSON 解析请求JSON
func ParseJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return errors.BadRequest("解析JSON失败: %s", err.Error())
	}
	return nil
}

// ParseQuery 解析Query参数
func ParseQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return errors.BadRequest("解析Query参数失败: %s", err.Error())
	}
	return nil
}

// ParseForm 解析表单
func ParseForm(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindWith(obj, binding.Form); err != nil {
		return errors.BadRequest("解析表单失败: %s", err.Error())
	}
	return nil
}

// GetToken 获取用户令牌
func GetToken(c *gin.Context) string {
	var token string
	auth := c.GetHeader("Authorization")
	prefix := "Bearer "
	if auth != "" && strings.HasPrefix(auth, prefix) {
		token = auth[len(prefix):]
	} else {
		token = auth
	}
	if token == "" {
		token = c.Query("accessToken")
	}
	return token
}

// GetBodyData 获取请求体数据
func GetBodyData(c *gin.Context) []byte {
	if v, ok := c.Get(ReqBodyKey); ok {
		if b, ok := v.([]byte); ok {
			return b
		}
	}
	return nil
}
