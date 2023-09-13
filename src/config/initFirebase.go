package config

import (
	"context"
	"log"
	"os"
	"sync"

	firebase "firebase.google.com/go/v4"

	"google.golang.org/api/option"
)

var app *firebase.App
var initOnce sync.Once

func InitializeApp() *firebase.App {
	initOnce.Do(func() {
		firebasePath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
		opt := option.WithCredentialsFile(firebasePath)
		application, err := firebase.NewApp(context.Background(), nil, opt)
		if err != nil {
			log.Fatal("error initializing firebase app:", err)
		}

		app = application
	})

	return app
}
