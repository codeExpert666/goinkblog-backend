package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"

	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
)

type CopyBodyConfig struct {
	AllowedPathPrefixes []string
	SkippedPathPrefixes []string
	MaxContentLen       int64
}

var DefaultCopyBodyConfig = CopyBodyConfig{
	MaxContentLen: 32 << 20, // 32MB
}

func CopyBody() gin.HandlerFunc {
	return CopyBodyWithConfig(DefaultCopyBodyConfig)
}

func CopyBodyWithConfig(config CopyBodyConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !AllowedPathPrefixes(c, config.AllowedPathPrefixes...) ||
			SkippedPathPrefixes(c, config.SkippedPathPrefixes...) ||
			c.Request.Body == nil {
			c.Next()
			return
		}

		var (
			requestBody []byte
			err         error
		)

		isGzip := false
		// http.MaxBytesReader 包装了 c.Request.Body，当已读取字节数超过 MaxContentLen 时，
		// 会：1、返回错误；2、向 c.Writer 写入 413 错误响应；3、关闭 c.Request.Body
		safe := http.MaxBytesReader(c.Writer, c.Request.Body, config.MaxContentLen)
		if c.GetHeader("Content-Encoding") == "gzip" {
			// 如果请求头中指定了 Content-Encoding 为 gzip，则解压请求体
			if reader, ierr := gzip.NewReader(safe); ierr == nil {
				isGzip = true
				// 读取解压后的请求体
				requestBody, err = io.ReadAll(reader)
			}
		}

		if !isGzip { // 未压缩，则直接读取请求体
			requestBody, err = io.ReadAll(safe)
		}

		if err != nil { // 请求体过大
			util.ResError(c, errors.RequestEntityTooLarge(fmt.Sprintf("请求体过大, 限制 %d 字节", config.MaxContentLen)))
			return
		}

		// 关闭原始请求体（释放底层网络资源）
		c.Request.Body.Close()
		// 创建一个新的内存缓冲区，包含复制的请求体内容（内存的好处是可重复读取）
		bf := bytes.NewBuffer(requestBody)
		// bf 是 io.Reader，io.NopCloser 将 io.Reader 转换为 io.ReadCloser，并且其 Close() 方法不执行任何操作
		// 由于 bf 是内存缓冲区，本身不需要关闭操作，恰好合适。转换出的 io.ReadCloser 满足 Body 类型，用于替换原始请求体为新的可重复读取的请求体
		c.Request.Body = io.NopCloser(bf)
		c.Set(util.ReqBodyKey, requestBody)
		c.Next()
	}
}
