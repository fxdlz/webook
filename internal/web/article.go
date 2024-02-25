package web

import (
	"github.com/gin-gonic/gin"
	"github.com/gotomicro/ekit/slice"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/internal/web/jwt"
	"webook/pkg/logger"
)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc service.InteractiveService
	log     logger.LoggerV1
	biz     string
}

func NewArticleHandler(svc service.ArticleService, intrSvc service.InteractiveService, log logger.LoggerV1) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		intrSvc: intrSvc,
		log:     log,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	g.POST("/edit", h.Edit)
	g.POST("/publish", h.Publish)
	g.POST("/withdraw", h.Withdraw)
	g.GET("/detail/:id", h.Detail)
	g.POST("/list", h.List)
	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	pub.POST("/like", h.Like)
	pub.POST("/collect", h.Collect)
}

func (h *ArticleHandler) Edit(ctx *gin.Context) {
	type Req struct {
		Id      int64
		Title   string
		Content string
	}
	req := Req{}
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		h.log.Error("保存文章数据失败", logger.Int64("uid", uc.Uid), logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (h *ArticleHandler) Publish(ctx *gin.Context) {
	type Req struct {
		Id      int64  `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}
	req := Req{}
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		h.log.Error("发表文章数据失败", logger.Int64("uid", uc.Uid), logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: id,
	})
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context) {
	type Req struct {
		Id int64 `json:"id"`
	}
	req := Req{}
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	err := h.svc.Withdraw(ctx, req.Id, uc.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg: "系统错误",
		})
		h.log.Error("撤回文章失败", logger.Int64("uid", uc.Uid), logger.Int64("id", req.Id), logger.Error(err))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.log.Warn("查询文章失败，id 格式不对",
			logger.String("id", idStr),
			logger.Error(err))
		return
	}

	art, err := h.svc.GetById(ctx, id)

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.log.Warn("查询文章失败",
			logger.Int64("id", id),
			logger.Error(err))
		return
	}

	uc := ctx.MustGet("user").(jwt.UserClaims)
	if uc.Uid != art.Author.Id {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.log.Error("非法查询文章",
			logger.Int64("id", id),
			logger.Int64("uid", uc.Uid))
		return
	}

	vo := ArticleVO{
		Id:       art.Id,
		Title:    art.Title,
		Abstract: art.Abstract(),
		Content:  art.Content,
		AuthorId: art.Author.Id,
		Status:   art.Status.ToUint8(),
		Ctime:    art.Ctime.Format(time.DateTime),
		Utime:    art.Utime.Format(time.DateTime),
	}

	ctx.JSON(http.StatusOK, Result{
		Data: vo,
	})
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}

	uc := ctx.MustGet("user").(jwt.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.Uid))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Data: slice.Map[domain.Article, ArticleVO](arts, func(idx int, src domain.Article) ArticleVO {
			return ArticleVO{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				AuthorId: src.Author.Id,
				Status:   src.Status.ToUint8(),
				Ctime:    src.Ctime.Format(time.DateTime),
				Utime:    src.Utime.Format(time.DateTime),
			}
		}),
	})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.log.Warn("查询文章失败，id 格式不对",
			logger.String("id", idStr),
			logger.Error(err))
		return
	}

	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)
	eg.Go(func() error {
		var er error
		art, er = h.svc.GetPubById(ctx, id)
		return er
	})

	uc := ctx.MustGet("user").(jwt.UserClaims)

	eg.Go(func() error {
		var er error
		intr, er = h.intrSvc.Get(ctx, h.biz, id, uc.Uid)
		return er
	})

	err = eg.Wait()

	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.log.Warn("查询文章失败",
			logger.Int64("id", id),
			logger.Error(err))
		return
	}

	err = h.intrSvc.IncrReadCnt(ctx, "", art.Id)
	if err != nil {
		//记录日志
		h.log.Error("更新阅读数失败", logger.Int64("aid", art.Id), logger.Error(err))
	}

	ctx.JSON(http.StatusOK, Result{
		Data: ArticleVO{
			Id:    art.Id,
			Title: art.Title,

			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			Status:     art.Status.ToUint8(),
			ReadCnt:    intr.ReadCnt,
			LikeCnt:    intr.LikeCnt,
			Liked:      intr.Liked,
			Collected:  intr.Collected,
			CollectCnt: intr.CollectCnt,
			Ctime:      art.Ctime.Format(time.DateTime),
			Utime:      art.Utime.Format(time.DateTime),
		},
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context) {
	type Req struct {
		Id   int64 `json:"id"`
		Like bool  `json:"like"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	var err error
	if req.Like {
		err = h.intrSvc.Like(ctx, h.biz, req.Id, uc.Uid)
	} else {
		err = h.intrSvc.CancelLike(ctx, h.biz, req.Id, uc.Uid)
	}
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("点赞/取消点赞失败",
			logger.Error(err),
			logger.Int64("uid", uc.Uid),
			logger.Int64("aid", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}

func (h *ArticleHandler) Collect(ctx *gin.Context) {
	type Req struct {
		Id  int64 `json:"id"`
		Cid int64 `json:"cid"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	var err error
	err = h.intrSvc.Collect(ctx, h.biz, req.Id, req.Cid, uc.Uid)
	if err != nil {
		ctx.JSON(http.StatusOK, Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.log.Error("收藏失败",
			logger.Error(err),
			logger.Int64("uid", uc.Uid),
			logger.Int64("aid", req.Id))
		return
	}
	ctx.JSON(http.StatusOK, Result{
		Msg: "OK",
	})
}
