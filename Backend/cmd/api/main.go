package main

import (
    "log"

    "radar-literario/backend/internal/router"
)

func main() {
    r := router.Novo()

    if err := r.Run("127.0.0.1:8080"); err != nil {
        log.Fatal(err)
    }
}