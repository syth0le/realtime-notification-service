package publicapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	xerrors "github.com/syth0le/gopnik/errors"
	"go.uber.org/zap"

	"github.com/go-http-utils/headers"

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

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		h.writeError(r.Context(), w, fmt.Errorf("cannot upgrade connection: %w", err))
		return
	}

	go func() {
		fmt.Println(ctx.Err(), <-ctx.Done())
		defer conn.Close()

		userID := model.UserID("ROUTING_KEY_USERID")

		err = h.notificationsService.SubscribeFeedNotifications(ctx, conn, &userID)
		if err != nil {
			h.logger.Sugar().Errorf("subscribe feed notifications: %v", err)
			return
		}
	}()

	// listen(conn)

}

func listen(conn net.Conn) {
	for {
		msg, op, err := wsutil.ReadClientData(conn)
		if err != nil {
			// handle error
		}
		err = wsutil.WriteServerMessage(conn, op, msg)
		if err != nil {
			// handle error
		}
	}
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

func writeResponse(w http.ResponseWriter, response any) {
	w.Header().Set(headers.ContentType, "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, xerrors.InternalErrorMessage, http.StatusInternalServerError) // TODO: make error mapping
	}
}
