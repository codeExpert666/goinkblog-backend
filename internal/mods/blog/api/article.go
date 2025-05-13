package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/biz"
	"github.com/codeExpert666/goinkblog-backend/internal/mods/blog/schema"
	"github.com/codeExpert666/goinkblog-backend/pkg/errors"
	"github.com/codeExpert666/goinkblog-backend/pkg/logging"
	"github.com/codeExpert666/goinkblog-backend/pkg/util"
)

// ArticleHandler 文章API处理器
type ArticleHandler struct {
	ArticleService *biz.ArticleService
}

// GetArticleList 获取（搜索）文章列表
// @Tags ArticleAPI
// @Summary 获取（搜索）文章列表
// @Param page query int false "页码" minimum(1) default(1)
// @Param page_size query int false "每页容量" minimum(1) maximum(100) default(10)
// @Param category_ids query []uint false "分类ID列表（可多选）" collectionFormat(multi) minimum(1)
// @Param tag_ids query []uint false "标签ID列表（可多选）" collectionFormat(multi) minimum(1)
// @Param author query string false "作者名称（current 表示当前用户）"
// @Param status query string false "状态" Enums(published, draft)
// @Param sort_by query string false "排序依据" Enums(newest, views, likes, favorites, comments) default(newest)
// @Param keyword query string false "搜索关键词"
// @Param time_range query string false "创建时间范围" Enums(today, week, month, year, all) default(all)
// @Success 200 {object} util.ResponseResult{data=schema.ArticlePaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles [get]
func (h *ArticleHandler) GetArticleList(c *gin.Context) {
	var params schema.ArticleQueryParams
	if err := util.ParseQuery(c, &params); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()

	// 处理作者参数
	if authorID := util.FromUserID(ctx); authorID == 0 && params.Author == "current" {
		util.ResError(c, errors.BadRequest("未登录用户无法获取当前用户的文章列表"))
		return
	}

	data, err := h.ArticleService.GetArticleList(ctx, &params)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Summary 获取文章详情
// @Param id path uint true "文章ID" minimum(1)
// @Success 200 {object} util.ResponseResult{data=schema.ArticleResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/{id} [get]
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的文章ID"))
		return
	}

	// 获取当前用户ID
	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)

	// 获取文章详情
	data, err := h.ArticleService.GetArticleByID(ctx, uint(id), userID)
	if err != nil {
		util.ResError(c, err)
		return
	}

	// 增加浏览次数（草稿不统计、作者不统计）
	if data.Status == "published" && data.AuthorID != userID {
		err = h.ArticleService.ViewArticle(ctx, userID, uint(id))
		if err != nil {
			if userID > 0 {
				logging.Context(ctx).Error("增加浏览次数失败", zap.Uint("article_id", uint(id)), zap.Uint("user_id", userID), zap.Error(err))
			} else {
				logging.Context(ctx).Error("增加浏览次数失败", zap.Uint("article_id", uint(id)), zap.String("user_id", "anonymous"), zap.Error(err))
			}
		} else {
			data.ViewCount++
		}
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 创建文章
// @Param title body string true "文章标题"
// @Param content body string true "文章内容"
// @Param summary body string false "文章摘要"
// @Param category_id body uint false "文章分类ID" minimum(1)
// @Param tag_ids body []uint false "文章标签ID列表"
// @Param cover body string false "文章封面图片URL"
// @Param status body string true "文章状态" enum("published", "draft")
// @Success 200 {object} util.ResponseResult{data=schema.ArticleResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles [post]
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var req schema.CreateArticleRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.ArticleService.CreateArticle(ctx, userID, &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 更新文章
// @Param id path uint true "文章ID"
// @Param title body string false "文章标题"
// @Param content body string false "文章内容"
// @Param summary body string false "文章摘要"
// @Param category_id body uint false "文章分类ID" minimum(1)
// @Param tag_ids body []uint false "文章标签ID列表"
// @Param cover body string false "文章封面图片URL"
// @Param status body string false "文章状态" enum("published", "draft")
// @Success 200 {object} util.ResponseResult{data=schema.ArticleResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 403 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/{id} [put]
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的文章ID"))
		return
	}

	var req schema.UpdateArticleRequest
	if err := util.ParseJSON(c, &req); err != nil {
		util.ResError(c, err)
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.ArticleService.UpdateArticle(ctx, userID, uint(id), &req)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 删除文章
// @Param id path uint true "文章ID"
// @Success 200 {object} util.ResponseResult
// @Failure 400 {object} util.ResponseResult
// @Failure 403 {object} util.ResponseResult
// @Failure 404 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/{id} [delete]
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的文章ID"))
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	err = h.ArticleService.DeleteArticle(ctx, userID, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResOK(c)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 上传文章封面
// @Accept multipart/form-data
// @Produce json
// @Param cover formData file true "文章封面图片"
// @Success 200 {object} util.ResponseResult{data=schema.CoverResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/upload-cover [post]
func (h *ArticleHandler) UploadCover(c *gin.Context) {
	data, err := h.ArticleService.UploadCover(c)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 点赞/取消点赞文章
// @Param id path uint true "文章ID"
// @Success 200 {object} util.ResponseResult{data=schema.ArticleInteractionResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/{id}/like [post]
func (h *ArticleHandler) LikeArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的文章ID"))
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.ArticleService.LikeArticle(ctx, userID, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 收藏/取消收藏文章
// @Param id path uint true "文章ID"
// @Success 200 {object} util.ResponseResult{data=schema.ArticleInteractionResponse}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/{id}/favorite [post]
func (h *ArticleHandler) FavoriteArticle(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的文章ID"))
		return
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.ArticleService.FavoriteArticle(ctx, userID, uint(id))
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 获取用户点赞的文章
// @Param page query int false "页数" minimum(1) default(1)
// @Param page_size query int false "页容量" minimum(1) maximum(100) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.ArticlePaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/liked [get]
func (h *ArticleHandler) GetUserLikedArticles(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页码"))
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页容量"))
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.ArticleService.GetUserLikedArticles(ctx, userID, page, pageSize)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 获取用户收藏的文章
// @Param page query int false "页数" minimum(1) default(1)
// @Param page_size query int false "页容量" minimum(1) maximum(100) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.ArticlePaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/favorites [get]
func (h *ArticleHandler) GetUserFavoriteArticles(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页码"))
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页容量"))
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.ArticleService.GetUserFavoriteArticles(ctx, userID, page, pageSize)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 获取用户浏览历史
// @Param page query int false "页数" minimum(1) default(1)
// @Param page_size query int false "页容量" minimum(1) maximum(100) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.ArticlePaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/history [get]
func (h *ArticleHandler) GetUserViewHistory(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页码"))
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页容量"))
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.ArticleService.GetUserViewHistory(ctx, userID, page, pageSize)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Security ApiKeyAuth
// @Summary 获取用户评论过的文章
// @Param page query int false "页数" minimum(1) default(1)
// @Param page_size query int false "页容量" minimum(1) maximum(100) default(10)
// @Success 200 {object} util.ResponseResult{data=schema.ArticlePaginationResult}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/commented [get]
func (h *ArticleHandler) GetUserCommentedArticles(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页码"))
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的页容量"))
	}

	ctx := c.Request.Context()
	userID := util.FromUserID(ctx)
	data, err := h.ArticleService.GetUserCommentedArticles(ctx, userID, page, pageSize)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Summary 获取热门文章
// @Param limit query int false "限制数量" default(5)
// @Success 200 {object} util.ResponseResult{data=[]schema.ArticleListItem}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/hot [get]
func (h *ArticleHandler) GetHotArticles(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的限制参数"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.ArticleService.GetHotArticles(ctx, limit)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}

// @Tags ArticleAPI
// @Summary 获取最新文章
// @Param limit query int false "限制数量" default(5)
// @Success 200 {object} util.ResponseResult{data=[]schema.ArticleListItem}
// @Failure 400 {object} util.ResponseResult
// @Failure 500 {object} util.ResponseResult
// @Router /api/blog/articles/latest [get]
func (h *ArticleHandler) GetLatestArticles(c *gin.Context) {
	limit, err := strconv.Atoi(c.DefaultQuery("limit", "5"))
	if err != nil {
		util.ResError(c, errors.BadRequest("无效的限制参数"))
		return
	}

	ctx := c.Request.Context()
	data, err := h.ArticleService.GetLatestArticles(ctx, limit)
	if err != nil {
		util.ResError(c, err)
		return
	}

	util.ResSuccess(c, data)
}
