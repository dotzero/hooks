package app

import (
	"context"
	"log"
	"time"
)

func (a *App) sweep(ctx context.Context) {
	for {
		select {
		case <-time.After(10 * time.Second):
			log.Printf("[DEBUG] sweep tick")

			maxAge := time.Duration(a.BoltTTL) * time.Hour

			if err := a.Storage.SweepHooks(maxAge); err != nil {
				log.Fatalf("[ERROR] failed to sweep hooks, %+v", err)
			}
		case <-ctx.Done():
			return
		}
	}
}
