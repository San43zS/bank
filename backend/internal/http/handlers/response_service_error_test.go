package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"banking-platform/internal/apperr"
	"github.com/gin-gonic/gin"
)

func TestRespondWithServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{})))

	testCases := []struct {
		name      string
		fullPath  string
		err       error
		wantCode  int
		wantError string
	}{
		{name: "public_bad_request", fullPath: "/x", err: apperr.BadRequest("bad_request"), wantCode: http.StatusBadRequest, wantError: "bad_request"},
		{name: "wrapped_public_bad_request", fullPath: "/x", err: fmt.Errorf("op: %w", apperr.BadRequest("bad_request")), wantCode: http.StatusBadRequest, wantError: "bad_request"},

		{name: "user_exists_conflict", fullPath: "/x", err: apperr.ErrUserExists, wantCode: http.StatusConflict, wantError: apperr.ErrUserExists.Error()},
		{name: "wrapped_user_exists_conflict", fullPath: "/x", err: fmt.Errorf("op: %w", apperr.ErrUserExists), wantCode: http.StatusConflict, wantError: apperr.ErrUserExists.Error()},

		{name: "invalid_credentials_unauthorized", fullPath: "/x", err: apperr.ErrInvalidCredentials, wantCode: http.StatusUnauthorized, wantError: apperr.ErrInvalidCredentials.Error()},
		{name: "invalid_token_unauthorized", fullPath: "/x", err: apperr.ErrInvalidToken, wantCode: http.StatusUnauthorized, wantError: apperr.ErrInvalidToken.Error()},
		{name: "unauthorized_forbidden", fullPath: "/x", err: apperr.ErrUnauthorized, wantCode: http.StatusForbidden, wantError: apperr.ErrUnauthorized.Error()},

		{name: "account_not_found", fullPath: "/x", err: apperr.ErrAccountNotFound, wantCode: http.StatusNotFound, wantError: apperr.ErrAccountNotFound.Error()},
		{name: "transaction_not_found", fullPath: "/x", err: apperr.ErrTransactionNotFound, wantCode: http.StatusNotFound, wantError: apperr.ErrTransactionNotFound.Error()},

		{name: "user_not_found_normal", fullPath: "/users/me", err: apperr.ErrUserNotFound, wantCode: http.StatusNotFound, wantError: apperr.ErrUserNotFound.Error()},
		{name: "user_not_found_transfer_is_400", fullPath: "/transactions/transfer", err: apperr.ErrUserNotFound, wantCode: http.StatusBadRequest, wantError: "recipient not found"},

		{name: "insufficient_funds", fullPath: "/x", err: apperr.ErrInsufficientFunds, wantCode: http.StatusBadRequest, wantError: apperr.ErrInsufficientFunds.Error()},
		{name: "invalid_amount", fullPath: "/x", err: apperr.ErrInvalidAmount, wantCode: http.StatusBadRequest, wantError: apperr.ErrInvalidAmount.Error()},
		{name: "invalid_currency", fullPath: "/x", err: apperr.ErrInvalidCurrency, wantCode: http.StatusBadRequest, wantError: apperr.ErrInvalidCurrency.Error()},
		{name: "currencies_must_differ", fullPath: "/x", err: apperr.ErrCurrenciesMustDiffer, wantCode: http.StatusBadRequest, wantError: apperr.ErrCurrenciesMustDiffer.Error()},
		{name: "cannot_transfer_to_self", fullPath: "/x", err: apperr.ErrCannotTransferToSelf, wantCode: http.StatusBadRequest, wantError: apperr.ErrCannotTransferToSelf.Error()},
		{name: "liquidity_unavailable_conflict", fullPath: "/x", err: apperr.ErrLiquidityUnavailable, wantCode: http.StatusConflict, wantError: apperr.ErrLiquidityUnavailable.Error()},

		{name: "unknown_internal", fullPath: "/x", err: errors.New("boom"), wantCode: http.StatusInternalServerError, wantError: "internal_error"},
		{name: "wrapped_unknown_internal", fullPath: "/x", err: fmt.Errorf("op: %w", errors.New("boom")), wantCode: http.StatusInternalServerError, wantError: "internal_error"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router := gin.New()
			router.GET(tc.fullPath, func(c *gin.Context) {
				respondWithServiceError(c, tc.err)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.fullPath, nil)
			router.ServeHTTP(w, req)

			if w.Code != tc.wantCode {
				t.Fatalf("status=%d want=%d body=%s", w.Code, tc.wantCode, w.Body.String())
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal=%v body=%s", err, w.Body.String())
			}
			if got, ok := body["error"].(string); !ok || got != tc.wantError {
				t.Fatalf("error=%v want=%v", body["error"], tc.wantError)
			}
		})
	}
}

