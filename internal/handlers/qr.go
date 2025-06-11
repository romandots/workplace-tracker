package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

// QRHandler generates a one-time QR code and stores the token in Redis.
func (e *Env) QRHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	token := uuid.NewString()

	ttl := 15 * time.Second
	if v := os.Getenv("QR_TTL"); v != "" {
		if dur, err := time.ParseDuration(v); err == nil {
			ttl = dur
		}
	}
	if err := e.App.Redis.Set(ctx, token, 1, ttl).Err(); err != nil {
		http.Error(w, "redis error", http.StatusInternalServerError)
		return
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		baseURL = fmt.Sprintf("%s://%s", scheme, r.Host)
	}
	url := fmt.Sprintf("%s/checkin?token=%s", baseURL, token)

	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "qr error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "no-store")
	w.Write(png)
}
