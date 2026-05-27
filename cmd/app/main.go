package main

import (
	"context"
	"support_chat/internal/app"
)

func main() {
	ctx := context.Background()
	a, err := app.New(ctx)
	if err != nil {
		panic(err)
	}
	a.Run()
}
