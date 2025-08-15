package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"go-ddd-architecture/app/domain/gametime"
	"go-ddd-architecture/app/infra/clock"
	"go-ddd-architecture/app/infra/memory"
	bb "go-ddd-architecture/app/infra/persistence/bbolt"
	"go-ddd-architecture/app/usecase/game"
)

var (
	flagDBPath    string
	flagUseMemory bool
)

var offlineCmd = &cobra.Command{
	Use:   "offline-claim",
	Short: "Initialize and claim offline gains, then print ViewModel",
	RunE: func(cmd *cobra.Command, args []string) error {
		clk := clock.SystemClock{}
		calc := gametime.NewOfflineCalculator()

		var (
			uc *game.Interactor
		)

		if flagUseMemory {
			m := memory.NewInMemoryRepo()
			uc = game.NewInteractor(m, clk, calc)
		} else {
			store, err := bb.New(flagDBPath)
			if err != nil {
				return err
			}
			defer store.Close()
			uc = game.NewInteractor(store, clk, calc)
		}

		if err := uc.Initialize(); err != nil {
			return err
		}

		res, err := uc.ClaimOffline(clk.Now())
		if err != nil {
			return err
		}
		vm := uc.GetViewModel()
		fmt.Printf("Offline Result: %+v\n", res)
		fmt.Printf("ViewModel: Knowledge=%d Research=%d\n", vm.Knowledge, vm.Research)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(offlineCmd)
	offlineCmd.Flags().StringVar(&flagDBPath, "db", "game.db", "path to bbolt db file")
	offlineCmd.Flags().BoolVar(&flagUseMemory, "mem", false, "use in-memory repository (no persistence)")
}
