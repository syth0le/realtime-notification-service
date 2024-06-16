package publicapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	xerrors "github.com/syth0le/gopnik/errors"
	"go.uber.org/zap"

	"github.com/go-http-utils/headers"

	"github.com/syth0le/realtime-notification-service/internal/clients/auth"
	"github.com/syth0le/realtime-notification-service/internal/model"
	"github.com/syth0le/realtime-notification-service/internal/service/notifications"
)

type Handler struct {
	logger               *zap.Logger
	notificationsService notifications.Service
}

func NewHandler(logger *zap.Logger, notificationsService notifications.Service) *Handler {
	return &Handler{
		logger:               logger,
		notificationsService: notificationsService,
	}
}

func (h *Handler) SubscribeFeedNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStr := ctx.Value(auth.UserIDValue)
	if userIDStr == "" {
		h.writeError(
			r.Context(),
			w,
			xerrors.WrapNotFoundError(fmt.Errorf("cannot recognize userID"), xerrors.NotFoundMessage),
		)
		return
	}
	userID := userIDStr.(model.UserID)

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		h.writeError(r.Context(), w, fmt.Errorf("cannot upgrade connection: %w", err))
		return
	}

	go func() {
		defer conn.Close()

		err = h.notificationsService.SubscribeFeedNotifications(ctx, conn, &userID)
		if err != nil {
			h.logger.Sugar().Errorf("subscribe feed notifications: %v", err)
			return
		}

		for {
			_, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				return
			}
			if op == ws.OpClose {
				// TODO: close connections and remove all pools
				return
			}
		}
	}()

}

func (h *Handler) writeError(ctx context.Context, w http.ResponseWriter, err error) {
	h.logger.Sugar().Warnf("http response error: %v", err)

	w.Header().Set(headers.ContentType, "application/json")
	errorResult, ok := xerrors.FromError(err)
	if !ok {
		h.logger.Sugar().Errorf("cannot write log message: %v", err)
		return
	}
	w.WriteHeader(errorResult.StatusCode)
	err = json.NewEncoder(w).Encode(
		map[string]any{
			"message": errorResult.Msg,
			"code":    errorResult.StatusCode,
		})

	if err != nil {
		http.Error(w, xerrors.InternalErrorMessage, http.StatusInternalServerError) // TODO: make error mapping
	}
}
