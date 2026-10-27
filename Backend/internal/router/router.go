package router

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "radar-literario/backend/internal/controller"
)

func Novo() *gin.Engine {
    r := gin.Default()

    r.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
        })
    })

    api := r.Group("/api")
    {
        api.GET("/usuarios/:id", controller.BuscarUsuario)
        api.POST("/usuarios", controller.CriarUsuario)
    }

    return r
}