package main

import (
	"io/fs"
	"log"
	"net/http"

	"childcare-backend/config"
	"childcare-backend/db"
	"childcare-backend/feishu"
	"childcare-backend/handler"
	"childcare-backend/middleware"
	"childcare-backend/store"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.Open(cfg.MYSQDSN)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer database.Close()

	migrations, err := fs.Sub(migrationsFS, "db/migrations")
	if err != nil {
		log.Fatalf("migrations fs: %v", err)
	}
	if err := db.RunMigrations(database, migrations, cfg.MYSQDSN); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// stores
	userStore    := store.NewUserStore(database)
	childStore   := store.NewChildStore(database)
	measureStore := store.NewMeasurementStore(database)
	sleepStore   := store.NewSleepStore(database)
	dietStore    := store.NewDietStore(database)
	suppStore    := store.NewSupplementStore(database)

	// handlers
	authH := handler.NewAuthHandler(userStore, cfg)
	childH := handler.NewChildHandler(childStore)
	measureH := handler.NewMeasurementHandler(measureStore, childStore)
	sleepH := handler.NewSleepHandler(sleepStore, childStore)
	dietH := handler.NewDietHandler(dietStore, childStore)
	supplementH := handler.NewSupplementHandler(suppStore, childStore)

	r := gin.Default()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authH.Register)
			auth.POST("/login",    authH.Login)
			auth.POST("/wx-login", authH.WxLogin)
			auth.POST("/refresh",  authH.Refresh)
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(cfg.JWTSecret))
		{
			withFamily := protected.Group("")
			{
				withFamily.GET("/children",                      childH.List)
				withFamily.POST("/children",                     childH.Create)
				withFamily.PUT("/children/:id",                  childH.Update)
				withFamily.DELETE("/children/:id",               childH.Delete)

				withFamily.GET("/children/:id/measurements",         measureH.List)
				withFamily.POST("/children/:id/measurements",        measureH.Create)
				withFamily.PUT("/children/:id/measurements/:mid",    measureH.Update)
				withFamily.DELETE("/children/:id/measurements/:mid", measureH.Delete)

				withFamily.GET("/children/:id/sleep",           sleepH.List)
				withFamily.POST("/children/:id/sleep",          sleepH.Create)
				withFamily.PUT("/children/:id/sleep/:sid",      sleepH.Update)
				withFamily.DELETE("/children/:id/sleep/:sid",   sleepH.Delete)

				withFamily.GET("/children/:id/diet",            dietH.List)
				withFamily.POST("/children/:id/diet",           dietH.Create)
				withFamily.PUT("/children/:id/diet/:did",       dietH.Update)
				withFamily.DELETE("/children/:id/diet/:did",    dietH.Delete)
				withFamily.GET("/children/:id/diet/types",      dietH.GetFoodTypes)

				withFamily.GET("/children/:id/supplements",           supplementH.List)
				withFamily.POST("/children/:id/supplements",          supplementH.Create)
				withFamily.PUT("/children/:id/supplements/:spid",     supplementH.Update)
				withFamily.DELETE("/children/:id/supplements/:spid",  supplementH.Delete)
				withFamily.GET("/children/:id/supplements/names",     supplementH.GetNames)

				withFamily.GET("/who-standards", handler.GetWHOStandards)
			}
		}
	}

	// Serve embedded frontend when built with -tags embed.
	if FrontendFS != nil {
		r.StaticFS("/assets", FrontendFS)
		r.GET("/favicon.svg", gin.WrapH(http.FileServer(FrontendFS)))
		r.GET("/icons.svg", gin.WrapH(http.FileServer(FrontendFS)))

		r.NoRoute(func(c *gin.Context) {
			indexFile, err := FrontendFS.Open("/index.html")
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			defer indexFile.Close()
			c.FileFromFS("/", FrontendFS)
		})
	}

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)

	// Start Feishu bot in background if configured.
	if cfg.FeishuAppID != "" && cfg.FeishuAppSecret != "" && cfg.DeepSeekAPIKey != "" {
		bot := feishu.NewBot(
			cfg.FeishuAppID, cfg.FeishuAppSecret, cfg.DeepSeekAPIKey,
			childStore, sleepStore, dietStore, suppStore, measureStore,
		)
		go func() {
			log.Printf("starting Feishu bot...")
			if err := bot.StartWS(); err != nil {
				log.Printf("feishu bot error: %v", err)
			}
		}()
	}

	if err := r.Run(addr); err != nil {
		log.Fatalf("run: %v", err)
	}
}
