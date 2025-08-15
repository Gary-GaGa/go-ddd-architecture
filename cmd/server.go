package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"go-ddd-architecture/app/adapter/in/httpserver"
	httpGame "go-ddd-architecture/app/adapter/in/httpserver/game"
	"go-ddd-architecture/app/domain/gametime"
	"go-ddd-architecture/app/infra/clock"
	"go-ddd-architecture/app/infra/memory"
	bb "go-ddd-architecture/app/infra/persistence/bbolt"
	"go-ddd-architecture/app/usecase/game"
	outPort "go-ddd-architecture/app/usecase/port/out/game"
)

// serverCmd -  represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Run: func(cmd *cobra.Command, args []string) {
		server()
	},
}

var (
	flagServerDBPath    string
	flagServerUseMemory bool
)

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.Flags().StringVar(&flagServerDBPath, "db", "game.db", "path to bbolt db file")
	serverCmd.Flags().BoolVar(&flagServerUseMemory, "mem", true, "use in-memory repository (no persistence)")
}

// server -
func server() {
	// DI
	app := fx.New(
		fx.NopLogger,
		// 基礎元件
		fx.Provide(
			// logger
			func() (*zap.Logger, error) { return zap.NewDevelopment() },
			func() clock.SystemClock { return clock.SystemClock{} },
			func() *gametime.OfflineCalculator { return gametime.NewOfflineCalculator() },
			// Repository：依旗標切換 memory 或 bbolt
			func() (outPort.Repository, error) {
				if flagServerUseMemory {
					return memory.NewInMemoryRepo(), nil
				}
				store, err := bb.New(flagServerDBPath)
				if err != nil {
					return nil, err
				}
				return store, nil
			},
			func(clk clock.SystemClock, calc *gametime.OfflineCalculator, repo outPort.Repository) *game.Interactor {
				return game.NewInteractor(repo, clk, calc)
			},
			// HTTP adapter
			// game 模組 handler/router
			func(uc *game.Interactor, log *zap.Logger) *httpGame.Handler { return httpGame.NewHandler(uc, log) },
			func(h *httpGame.Handler) *httpGame.Router { return httpGame.NewRouter(h) },
			// 聚合 router
			func(gr *httpGame.Router) *httpserver.Router { return httpserver.NewRouter(gr) },
			func(r *httpserver.Router, log *zap.Logger) *httpserver.Server {
				return httpserver.NewServer("127.0.0.1:8080", r.HandlerWithLogger(log))
			},
		),
		fx.Invoke(NewGRPCServer, StartHTTPServer, InitUsecase),
	)

	if err := app.Err(); err != nil {
		log.Fatal(err)
	}

	app.Run()
}

// NewGRPCServer -
func NewGRPCServer(lc fx.Lifecycle) error {

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {

			return nil
		},
		OnStop: func(ctx context.Context) error {

			return nil
		},
	})

	return nil
}

// StartHTTPServer 啟動 HTTP 伺服器
func StartHTTPServer(lc fx.Lifecycle, srv *httpserver.Server) error {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return srv.Start() },
		OnStop:  func(ctx context.Context) error { return srv.Shutdown(ctx) },
	})
	return nil
}

// InitUsecase 在啟動時載入資料到 Interactor，避免初次請求時使用零值 timestamps。
func InitUsecase(lc fx.Lifecycle, uc *game.Interactor) error {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error { return uc.Initialize() },
	})
	return nil
}
