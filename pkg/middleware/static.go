package middleware

import (
	"os"
	"path/filepath"

	"github.com/codeExpert666/goinkblog-backend/pkg/util"
	"github.com/gin-gonic/gin"
)

type StaticConfig struct {
	SkippedPathPrefixes []string
	// 静态文件根目录
	Root string
}

func StaticWithConfig(config StaticConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if SkippedPathPrefixes(c, config.SkippedPathPrefixes...) {
			c.Next()
			return
		}

		p := c.Request.URL.Path
		if p == "/" { // 首页重定向到 index.html
			p = "index.html"
		}

		// filepath.FromSlash 用于将URL中的斜杠转换为系统对应的路径分隔符
		fpath := filepath.Join(config.Root, filepath.FromSlash(p))

		fileInfo, err := os.Stat(fpath)
		if (err != nil && os.IsNotExist(err)) || fileInfo.IsDir() { // 文件不存在或为目录
			fpath = filepath.Join(config.Root, "index.html")
		}

		if !util.IsImageFile(fpath) && !util.IsFrontendFile(fpath) { // 非图片文件、非前端文件支持下载
			// 设置下载相关的响应头
			fileName := filepath.Base(fpath)
			c.Header("Content-Disposition", "attachment; filename="+fileName) // 指示浏览器将响应内容作为附件下载，并指定文件名
			c.Header("Content-Description", "File Transfer")                  // 提供文件传输的描述
			c.Header("Content-Transfer-Encoding", "binary")                   // 设置为二进制传输
			c.Header("Content-Type", "application/octet-stream")              //设置为通用的二进制流类型
		}
		c.File(fpath)
		c.Abort()
	}
}
