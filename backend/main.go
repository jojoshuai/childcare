package main

import (
	"log"

	"childcare-backend/config"
	"childcare-backend/db"
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

	if err := db.RunMigrations(database, "file://db/migrations"); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	// stores
	userStore    := store.NewUserStore(database)
	familyStore  := store.NewFamilyStore(database)
	childStore   := store.NewChildStore(database)
	measureStore := store.NewMeasurementStore(database)
	inviteStore  := store.NewInviteStore(database)

	// handlers
	authH    := handler.NewAuthHandler(userStore, familyStore, cfg)
	familyH  := handler.NewFamilyHandler(familyStore, userStore, inviteStore)
	childH   := handler.NewChildHandler(childStore)
	measureH := handler.NewMeasurementHandler(measureStore, childStore)

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
			protected.GET("/family",        familyH.GetFamily)
			protected.POST("/family/invite", familyH.GenerateInvite)
			protected.POST("/family/join",   familyH.JoinFamily)

			withFamily := protected.Group("")
			withFamily.Use(middleware.RequireFamily())
			{
				withFamily.GET("/children",                          childH.List)
				withFamily.POST("/children",                         childH.Create)
				withFamily.PUT("/children/:id",                      childH.Update)
				withFamily.DELETE("/children/:id",                   childH.Delete)

				withFamily.GET("/children/:id/measurements",             measureH.List)
				withFamily.POST("/children/:id/measurements",            measureH.Create)
				withFamily.PUT("/children/:id/measurements/:mid",        measureH.Update)
				withFamily.DELETE("/children/:id/measurements/:mid",     measureH.Delete)

				withFamily.GET("/who-standards", handler.GetWHOStandards)
			}
		}
	}

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("run: %v", err)
	}
}
