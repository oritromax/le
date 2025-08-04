package logger

import (
	"context"
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"go.sakib.dev/le/pkg/utils"
)


const (
	StatusCodeKey string = "statusCode"
)

type Handler struct {
	slog.Handler
}

func NewHandler() *Handler {
	return &Handler{
		Handler: tint.NewHandler(
			os.Stdout,
			&tint.Options{
				Level: slog.LevelInfo,
			},
		),
	}
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	reqId, ok := ctx.Value(utils.RequestIDKey).(string)

	if !ok {
		return h.Handler.Handle(ctx, r)
	}

	r.AddAttrs(slog.String(string(utils.RequestIDKey), reqId))

	return h.Handler.Handle(ctx, r)
}
