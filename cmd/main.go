package main

import (
	"context"

	"github.com/Badchaos11/TSU_TT/config"
	"github.com/Badchaos11/TSU_TT/service"
)

func main() {
	ctx := context.Background()
	conf, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	app, err := service.NewService(ctx, conf)
	if err != nil {
		panic(err)
	}

	app.Run()
}
