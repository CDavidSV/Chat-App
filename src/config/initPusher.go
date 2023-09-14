package config

import (
	"sync"

	pusher "github.com/pusher/pusher-http-go/v5"
)

var pusherClient *pusher.Client
var pusherOnce sync.Once

func PusherInit() *pusher.Client {
	pusherOnce.Do(func() {
		client := pusher.Client{
			AppID:   "1670925",
			Key:     "ee7ae47591298cf84395",
			Secret:  "eaedece15925078cf9c7",
			Cluster: "mt1",
			Secure:  true,
		}

		pusherClient = &client
	})

	return pusherClient
}
