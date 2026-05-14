package router

import (
	"shopapi/internal/handler"
	"shopapi/internal/middleware"
	"shopapi/internal/service"

	"github.com/gin-gonic/gin"
)

// New wires HTTP routes: public catalog, JWT-protected mutations, auth, and orders.
func New(
	auth *service.AuthService,
	authH *handler.AuthHandler,
	productH *handler.ProductHandler,
	orderH *handler.OrderHandler,
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", authH.Register)
		authGroup.POST("/login", authH.Login)
		authGroup.GET("/me", middleware.JWTAuth(auth), authH.Me)
	}

	r.GET("/products", productH.List)
	r.GET("/products/:id", productH.Get)

	protected := r.Group("")
	protected.Use(middleware.JWTAuth(auth))
	{
		protected.POST("/products", productH.Create)
		protected.PUT("/products/:id", productH.Update)
		protected.DELETE("/products/:id", productH.Delete)

		protected.GET("/orders", orderH.List)
		protected.POST("/orders", orderH.Place)
	}

	return r
}
