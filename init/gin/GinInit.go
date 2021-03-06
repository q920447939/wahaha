package gin

import (
	"github.com/gin-gonic/gin"
	config "wahaha/conf"
	"runtime"
	"github.com/gin-contrib/pprof"
			"net/http"
	"wahaha/module/logger"
	"github.com/go-sql-driver/mysql"
	ig "wahaha/module/gin"
	"github.com/gin-gonic/contrib/sessions"
)

func InitGin(r *gin.Engine) {
	ig.GinEngine = r
	runtime.GOMAXPROCS(runtime.NumCPU())
	if config.GetEnv().ServerConfig.DEBUG {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	if config.GetEnv().ServerConfig.DEBUG {
		pprof.Register(ig.GinEngine) // 性能分析工具
	}

	loadHtmlGlob()

	ginProperties()
}

func ginProperties() {
	//ig.GinEngine.Use(gin.Logger())

	ig.GinEngine.Use(handleErrors())            // 错误处理
	store := sessions.NewCookieStore([]byte("secret"))
	ig.GinEngine.Use(sessions.Sessions("mysession", store))

/*	ig.GinEngine.Use(filters.RegisterSession()) // 全局session
	ig.GinEngine.Use(filters.RegisterCache())   // 全局cache
	ig.GinEngine.Use(auth.RegisterGlobalAuthDriver("cookie", "web_auth")) // 全局auth cookie
	ig.GinEngine.Use(auth.RegisterGlobalAuthDriver("jwt", "jwt_auth"))    // 全局auth jwt*/

	ig.GinEngine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "找不到该路由",
		})
		return
	})

	ig.GinEngine.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{
			"code": 404,
			"msg":  "找不到该方法",
		})
		return
	})
}

func loadHtmlGlob() {
	ig.GinEngine.LoadHTMLGlob("views/*") // html模板
	//ig.GinEngine.LoadHTMLGlob("views/**/*")
}

func handleErrors() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {

				logger.Error(err)
				panic(err)

				var (
					errMsg     string
					mysqlError *mysql.MySQLError
					ok         bool
				)
				if errMsg, ok = err.(string); ok {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code": 500,
						"msg":  "system error, " + errMsg,
					})
					return
				} else if mysqlError, ok = err.(*mysql.MySQLError); ok {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code": 501,
						"msg":  "system error, " + mysqlError.Error(),
					})
					return
				} else {
					c.JSON(http.StatusInternalServerError, gin.H{
						"code": 502,
						"msg":  "system error",
					})
					return
				}
			}
		}()
		c.Next()
	}
}
