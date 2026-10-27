package controller

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func BuscarUsuario(c *gin.Context) {
    // Obtém o parâmetro presente na URL.
    id := c.Param("id")

    c.JSON(http.StatusOK, gin.H{
        "id":   id,
        "nome": "Usuário de teste",
    })
}

func CriarUsuario(c *gin.Context) {
    // Estrutura esperada no corpo JSON da requisição.
    var entrada struct {
        Nome  string `json:"nome" binding:"required"`
        Email string `json:"email" binding:"required,email"`
    }

    if err := c.ShouldBindJSON(&entrada); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "erro": "Informe nome e um email válido",
        })
        return
    }

    // Apenas devolve os dados recebidos; não salva.
    c.JSON(http.StatusOK, gin.H{
        "mensagem": "Dados recebidos, sem persistência",
        "usuario":  entrada,
    })
}