package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"lang-learn-bot/telegramapi"
	"log/slog"
	"net/http"
)

func ListenForUpdates(ctx context.Context, updateChan chan telegramapi.Update, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("incoming request")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.Error("reading request body failed: " + err.Error())
			return
		}
		var update telegramapi.Update
		err = json.Unmarshal(body, &update)
		if err != nil {
			slog.Error("json error: " + err.Error())
		}
		updateChan <- update
		fmt.Fprintf(w, "ok")
	})
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	go func() {
		slog.Info("listening for webhook updates on port 8080")
		if err := srv.ListenAndServe(); err != nil {
			slog.Error("server error: " + err.Error())
		}
	}()
	<-ctx.Done()
	srv.Shutdown(ctx)
}
