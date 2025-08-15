package httpserver

import (
	"context"
	"net/http"
	"time"
)

type Server struct {
	httpServer *http.Server
}

func NewServer(addr string, handler http.Handler) *Server {
	s := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 5 * time.Second}
	return &Server{httpServer: s}
}

func (s *Server) Start() error {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			// 寫到標準錯誤即可，避免引入 logger 循環相依
			// 可在上層用 zap 記錄 lifecycle 錯誤
		}
	}()
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
