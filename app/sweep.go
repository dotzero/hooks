package app

import (
	"context"
	"log"
	"time"

	"github.com/dotzero/hooks/app/storage"
)

func (a *App) sweep(ctx context.Context) {
	for {
		select {
		case <-time.After(10 * time.Second):
			log.Printf("[DEBUG] sweep tick")

			maxAge := time.Duration(a.BoltTTL) * time.Hour

			if err := a.Storage.Sweep(storage.BucketHooks, storage.BucketHooksTTL, maxAge); err != nil {
				log.Fatalf("[ERROR] failed to sweep hooks, %+v", err)
			}

			if err := a.Storage.Sweep(storage.BucketReqs, storage.BucketReqsTTL, maxAge); err != nil {
				log.Fatalf("[ERROR] failed to sweep requests, %+v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
