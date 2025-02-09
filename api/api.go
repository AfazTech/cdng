package api

import (
	"log"
	"net/http"

	"github.com/imafaz/cdng/controller"

	"github.com/gin-gonic/gin"
)

func authMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("apiKey", apiKey)
		key := c.GetHeader("Authorization")
		if key != "Bearer "+apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "message": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func SetupRouter(apiKey string) *gin.Engine {
	r := gin.Default()
	r.Use(authMiddleware(apiKey))

	r.POST("/add-domain", func(c *gin.Context) {
		var req struct {
			Domain string `json:"domain"`
			IP     string `json:"ip"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": err.Error()})
			return
		}
		if err := controller.AddDomain(req.Domain, req.IP); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Domain added successfully"})
	})

	r.DELETE("/delete-domain/:domain", func(c *gin.Context) {
		domain := c.Param("domain")
		if err := controller.DeleteDomain(domain); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Domain deleted successfully"})
	})

	r.POST("/add-port", func(c *gin.Context) {
		var req struct {
			Port string `json:"port"`
		}
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": err.Error()})
			return
		}
		if err := controller.AddPort(req.Port); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Port added successfully"})
	})

	r.DELETE("/delete-port/:port", func(c *gin.Context) {
		port := c.Param("port")
		if err := controller.DeletePort(port); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Port deleted successfully"})
	})

	r.GET("/status", func(c *gin.Context) {
		status, err := controller.StatusNginx()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "status": status})
	})

	r.POST("/reload", func(c *gin.Context) {
		if err := controller.ReloadNginx(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Nginx reloaded successfully"})
	})

	r.POST("/stop", func(c *gin.Context) {
		if err := controller.StopNginx(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Nginx stopped successfully"})
	})

	r.POST("/restart", func(c *gin.Context) {
		if err := controller.RestartNginx(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "message": "Nginx restarted successfully"})
	})

	r.GET("/stats", func(c *gin.Context) {
		stats, err := controller.GetStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "stats": stats})
	})

	return r
}

func StartServer(port string, apiKey string) {
	r := SetupRouter(apiKey)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
