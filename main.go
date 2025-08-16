package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/iawia002/lux/extractors"
	"github.com/iawia002/lux/extractors/bilibili"
	"github.com/iawia002/lux/extractors/douyin"
	"github.com/iawia002/lux/extractors/douyu"
	"github.com/iawia002/lux/extractors/eporner"
	"github.com/iawia002/lux/extractors/facebook"
	"github.com/iawia002/lux/extractors/haokan"
	"github.com/iawia002/lux/extractors/huya"
	"github.com/iawia002/lux/extractors/instagram"
	"github.com/iawia002/lux/extractors/kuaishou"
	"github.com/iawia002/lux/extractors/mgtv"
	"github.com/iawia002/lux/extractors/netease"
	"github.com/iawia002/lux/extractors/pinterest"
	"github.com/iawia002/lux/extractors/pixivision"
	"github.com/iawia002/lux/extractors/pornhub"
	"github.com/iawia002/lux/extractors/qq"
	"github.com/iawia002/lux/extractors/tiktok"
	"github.com/iawia002/lux/extractors/tumblr"
	"github.com/iawia002/lux/extractors/twitter"
	"github.com/iawia002/lux/extractors/universal"
	"github.com/iawia002/lux/extractors/weibo"
	"github.com/iawia002/lux/extractors/xiaohongshu"
	"github.com/iawia002/lux/extractors/xvideos"
	"github.com/iawia002/lux/extractors/youtube"
	"github.com/iawia002/lux/extractors/youku"
)

func main() {
	extractors.Register("bilibili", bilibili.New())
	extractors.Register("douyin", douyin.New())
	extractors.Register("douyu", douyu.New())
	extractors.Register("eporner", eporner.New())
	extractors.Register("facebook", facebook.New())
	extractors.Register("haokan", haokan.New())
	extractors.Register("huya", huya.New())
	extractors.Register("instagram", instagram.New())
	extractors.Register("kuaishou", kuaishou.New())
	extractors.Register("mgtv", mgtv.New())
	extractors.Register("netease", netease.New())
	extractors.Register("pinterest", pinterest.New())
	extractors.Register("pixivision", pixivision.New())
	extractors.Register("pornhub", pornhub.New())
	extractors.Register("qq", qq.New())
	extractors.Register("tiktok", tiktok.New())
	extractors.Register("tumblr", tumblr.New())
	extractors.Register("twitter", twitter.New())
	extractors.Register("universal", universal.New())
	extractors.Register("weibo", weibo.New())
	extractors.Register("xiaohongshu", xiaohongshu.New())
	extractors.Register("xvideos", xvideos.New())
	extractors.Register("youtube", youtube.New())
	extractors.Register("youku", youku.New())
	r := gin.Default()
	api := r.Group("/api")
	{
		api.GET("/extract", func(c *gin.Context) {
			url := c.Query("url")
			if url == "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "url is required",
				})
				return
			}

			data, err := extractors.Extract(url, extractors.Options{})
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, data)
		})
	}

	r.Run(":8080") // listen and serve on 0.0.0.0:8080
}
