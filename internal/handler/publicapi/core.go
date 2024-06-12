package publicapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	xerrors "github.com/syth0le/gopnik/errors"
	"go.uber.org/zap"

	"github.com/go-http-utils/headers"

	"github.com/syth0le/realtime-notification-service/internal/service/notifications"
)

type Handler struct {
	logger               *zap.Logger
	upgrader             websocket.Upgrader
	notificationsService notifications.Service
}

func NewHandler(logger *zap.Logger, notificationsService notifications.Service) *Handler {
	return &Handler{
		logger: logger,
		upgrader: websocket.Upgrader{
			HandshakeTimeout:  0,
			ReadBufferSize:    1024,
			WriteBufferSize:   1024,
			WriteBufferPool:   nil,
			Subprotocols:      nil,
			Error:             nil,
			CheckOrigin:       nil,
			EnableCompression: false, // todo: move to gopnik with cfg
		},
		notificationsService: notificationsService,
	}
}

func (h *Handler) SubscribeFeedNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Sugar().Errorf("cannot upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// listen(conn)

	err = h.notificationsService.SubscribeFeedNotifications(ctx, conn, "ROUTING_KEY_USERID")
	if err != nil {
		h.logger.Sugar().Errorf("subscribe feed notifications: %v", err)
		return
	}
}

func listen(conn *websocket.Conn) {
	for {
		// read a message
		messageType, messageContent, err := conn.ReadMessage()
		timeReceive := time.Now()
		if err != nil {
			log.Println(err)
			return
		}

		// print out that message
		fmt.Println(string(messageContent))

		// reponse message
		messageResponse := fmt.Sprintf("Your message is: %s. Time received : %v", messageContent, timeReceive)

		if err := conn.WriteMessage(messageType, []byte(messageResponse)); err != nil {
			log.Println(err)
			return
		}
	}
}

func writeResponse(w http.ResponseWriter, response any) {
	w.Header().Set(headers.ContentType, "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, xerrors.InternalErrorMessage, http.StatusInternalServerError) // TODO: make error mapping
	}
}
