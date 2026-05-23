package scheduler

import (
	"context"
	"log"
	"time"

	verificationsvciface "github.com/sunibgram/domain/service/verification"
)

const verificationCodePurgeInterval = 1 * time.Hour


func RunVerificationCodePurge(ctx context.Context, svc verificationsvciface.IVerificationService) {
	ticker := time.NewTicker(verificationCodePurgeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if n, err := svc.PurgeExpiredCodes(); err != nil {
				log.Printf("PurgeExpiredCodes error: %v", err)
			} else if n > 0 {
				log.Printf("verification_codes purge: %d rows", n)
			}
		}
	}
}
