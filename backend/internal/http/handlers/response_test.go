package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRespondWithBindError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type emailReq struct {
		Email string `json:"email" binding:"required,email"`
	}
	type amountReq struct {
		AmountCents int64 `json:"amount_cents" binding:"required,gt=0"`
	}
	type registerReq struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}

	tests := []struct {
		name       string
		makeErr    func() error
		wantStatus int
		wantError  any
		wantFields bool
	}{
		{
			name: "empty_body",
			makeErr: func() error {
				return io.EOF
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "request body is required",
			wantFields: false,
		},
		{
			name: "invalid_json",
			makeErr: func() error {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"email":`))
				c.Request.Header.Set("Content-Type", "application/json")
				var r emailReq
				return c.ShouldBindJSON(&r)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid json",
			wantFields: false,
		},
		{
			name: "invalid_type",
			makeErr: func() error {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"amount_cents":"x"}`))
				c.Request.Header.Set("Content-Type", "application/json")
				var r amountReq
				return c.ShouldBindJSON(&r)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "invalid field type: amount_cents",
			wantFields: false,
		},
		{
			name: "validation_email",
			makeErr: func() error {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"email":"not-an-email"}`))
				c.Request.Header.Set("Content-Type", "application/json")
				var r emailReq
				return c.ShouldBindJSON(&r)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
			wantFields: true,
		},
		{
			name: "validation_min",
			makeErr: func() error {
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"email":"a@b.com","password":"123"}`))
				c.Request.Header.Set("Content-Type", "application/json")
				var r registerReq
				return c.ShouldBindJSON(&r)
			},
			wantStatus: http.StatusBadRequest,
			wantError:  "validation_error",
			wantFields: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			respondWithBindError(c, tt.makeErr())

			if w.Code != tt.wantStatus {
				t.Fatalf("status=%d want=%d", w.Code, tt.wantStatus)
			}

			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("unmarshal=%v", err)
			}
			if body["error"] != tt.wantError {
				t.Fatalf("error=%v want=%v", body["error"], tt.wantError)
			}

			_, hasFields := body["fields"]
			if hasFields != tt.wantFields {
				t.Fatalf("hasFields=%v want=%v", hasFields, tt.wantFields)
			}
		})
	}
}

