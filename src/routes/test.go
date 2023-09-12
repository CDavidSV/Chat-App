package routes

import (
	"net/http"
	"fmt"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("test!!!!"))
	fmt.Println("test!!!!")
}