package main

import (
	"context"
	"time"
)

func main() {
	client := NewClient()
	ctx := context.Background()
	poller(ctx, 2*time.Second, 10*time.Second, client)
}
