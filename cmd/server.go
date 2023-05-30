package cmd

import (
	"context"
	"log"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

// serverCmd -  represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "start server",
	Run: func(cmd *cobra.Command, args []string) {
		server()
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

// server -
func server() {
	// DI
	app := fx.New(
		fx.NopLogger,
		fx.Provide(),
		fx.Invoke(NewGRPCServer),
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
