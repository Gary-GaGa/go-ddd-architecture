package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Middleware 定義用於鏈式包裝 http.Handler 的函式型別。
type Middleware func(http.Handler) http.Handler

// Chain 依序套用多個 middleware 至處理器。
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

type ctxKey string

const ctxKeyRequestID ctxKey = "req_id"
const headerRequestID = "X-Request-Id"

// RequestIDMiddleware 產生或沿用請求 ID，並放入 Header 與 Context。
func RequestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(headerRequestID)
		if rid == "" {
			rid = randomID()
		}
		w.Header().Set(headerRequestID, rid)
		ctx := context.WithValue(r.Context(), ctxKeyRequestID, rid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RecoveryMiddleware 捕捉 panic，回傳 500 並記錄日誌。
func NewRecoveryMiddleware(log *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					rid, _ := r.Context().Value(ctxKeyRequestID).(string)
					log.Error("panic recovered", zap.Any("panic", rec), zap.String("rid", rid))
					if rw, ok := w.(*statusRecorder); ok {
						if rw.status == 0 {
							writeJSONError(rw, http.StatusInternalServerError, "internal", "internal server error")
							return
						}
					}
					writeJSONError(w, http.StatusInternalServerError, "internal", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// AccessLogMiddleware 記錄請求摘要（方法、路徑、狀態碼、耗時、request id）。
func NewAccessLogMiddleware(log *zap.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sr := &statusRecorder{ResponseWriter: w}
			next.ServeHTTP(sr, r)
			dur := time.Since(start)
			rid, _ := r.Context().Value(ctxKeyRequestID).(string)
			log.Info("http_request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", sr.statusOrDefault()),
				zap.Duration("duration", dur),
				zap.String("rid", rid),
			)
		})
	}
}

// statusRecorder 用於攔截狀態碼

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(statusCode int) {
	s.status = statusCode
	s.ResponseWriter.WriteHeader(statusCode)
}

func (s *statusRecorder) statusOrDefault() int {
	if s.status == 0 {
		return http.StatusOK
	}
	return s.status
}

func randomID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// 極少數情況下失敗，退回時間字串
		return time.Now().UTC().Format("20060102150405.000000")
	}
	return hex.EncodeToString(b[:])
}

// 輕量錯誤 JSON 寫入，避免 import 造成循環
type errEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSONError(w http.ResponseWriter, status int, code, msg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	var e errEnvelope
	e.Error.Code = code
	e.Error.Message = msg
	_ = json.NewEncoder(w).Encode(e)
}
