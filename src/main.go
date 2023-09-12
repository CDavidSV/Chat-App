package main

import (
    "net/http"
    "chat-app-back/src/routes"
)

func main() {
    http.HandleFunc("/test", routes.TestHandler)

    http.ListenAndServe(":8080", nil)
}