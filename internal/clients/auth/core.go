package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-http-utils/headers"
	xerrors "github.com/syth0le/gopnik/errors"
	inpb "github.com/syth0le/social-network/pkg/proto/internalapi"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const authHeader = "Authorization"
const UserIDValue = "userID"

type Client interface {
	AuthenticationInterceptor(next http.Handler) http.Handler
}

type ClientImpl struct {
	logger *zap.Logger
	client inpb.AuthServiceClient
}

func NewAuthImpl(logger *zap.Logger, conn *grpc.ClientConn) *ClientImpl {
	return &ClientImpl{
		logger: logger,
		client: inpb.NewAuthServiceClient(conn),
	}
}

func (c *ClientImpl) AuthenticationInterceptor(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authToken := r.Header.Get(authHeader)

		userID, err := c.client.ValidateToken(r.Context(), &inpb.ValidateTokenRequest{Token: authToken})
		if err != nil {
			c.writeError(w, fmt.Errorf("validate token: %w", err))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDValue, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (c *ClientImpl) writeError(w http.ResponseWriter, err error) {
	c.logger.Sugar().Warnf("http response error: %v", err)

	w.Header().Set(headers.ContentType, "application/json")
	errorResult, ok := xerrors.FromError(err)
	if !ok {
		c.logger.Sugar().Errorf("cannot write log message: %v", err)
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
