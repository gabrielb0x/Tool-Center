package main

import (
	"fmt"
	"log"
	"os"

	"toolcenter/config"
	"toolcenter/scripts/auth"
	"toolcenter/scripts/tools"
	"toolcenter/scripts/user"
	"toolcenter/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
)

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "https://tool-center.fr")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func watchConfig(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("❌ Erreur watcher:", err)
	}
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op == fsnotify.Write {
					log.Println("⚠️ config.json modifié → reload")
					if err := config.Load(path); err != nil {
						log.Println("❌ Erreur reload config :", err)
					} else {
						log.Println("✅ Configuration rechargée.")
					}
				}
			case err := <-watcher.Errors:
				log.Println("❌ Erreur watcher :", err)
			}
		}
	}()
	if err := watcher.Add(path); err != nil {
		log.Fatal("❌ Erreur ajout watcher :", err)
	}
}

func getDir(path string) string {
	lastSlash := -1
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			lastSlash = i
			break
		}
	}
	if lastSlash == -1 {
		return "."
	}
	return path[:lastSlash]
}

func setupLogger() {
	cfg := config.Get()

	if cfg.Logs.Enabled {
		logPath := cfg.Logs.Path
		if logPath == "" {
			log.Println("⚠️ Aucun chemin de log spécifié dans la config")
			return
		}

		dir := getDir(logPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				log.Fatal("❌ Erreur création dossier logs :", err)
			}
		}

		file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal("❌ Erreur ouverture fichier log :", err)
		}
		log.SetOutput(file)
	}
}

func setupRoutes(r *gin.Engine) {
	cfg := config.Get()
	api := r.Group("/" + cfg.URLVersion)

	api.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": fmt.Sprintf("ToolCenter API v%s is running", cfg.Version),
		})
	})

	authGroup := api.Group("/auth")
	authGroup.POST("/login", auth.LoginHandler)
	authGroup.POST("/register", auth.RegisterHandler)

	userGroup := api.Group("/user")
	userGroup.GET("/verify_email", user.VerifyEmailHandler)
	userGroup.GET(("/me"), user.MeHandler)
	userGroup.POST("/avatar", user.UploadAvatar)

	toolsGroup := api.Group("/tools")
	toolsGroup.POST("/add", tools.SubmitToolHandler)
	toolsGroup.GET("/me", tools.MyToolsHandler)

	utilsGroup := api.Group("/utils")
	utilsGroup.POST("/privates_news", utils.PrivateNewsHandler)
	utilsGroup.POST("/discord_webhook", utils.DiscordWebhookNewsHandler)
}

func main() {
	cfgPath := "config.json"

	if err := config.Load(cfgPath); err != nil {
		log.Fatal("❌ Erreur chargement config :", err)
	}

	setupLogger()
	go watchConfig(cfgPath)

	cfg := config.Get()
	if cfg.GinMode != "" {
		gin.SetMode(cfg.GinMode)
	}

	r := gin.Default()
	r.Use(corsMiddleware())
	setupRoutes(r)

	log.Printf("✅ API ToolCenter démarrée sur le port %d", cfg.Port)
	r.Run(fmt.Sprintf(":%d", cfg.Port))
}
