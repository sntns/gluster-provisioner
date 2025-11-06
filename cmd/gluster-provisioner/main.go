package main

import (
	"fmt"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/fx"
)

var path = "./config"

func Run(options ...fx.Option) error {
	app := fx.New(options...)
	c := app.Wait()
	app.Run()
	if err := app.Err(); err != nil {
		return err
	}
	if sig, ok := <-c; ok {
		if sig.ExitCode != 0 {
			os.Exit(sig.ExitCode)
		} else if ssig, ok := sig.Signal.(syscall.Signal); ok {
			os.Exit(int(ssig))
		}
		return fmt.Errorf("closing signal: %v", sig.Signal)
	}
	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "gluster-provisioner",
		Short: "Gluster Provisioner",
	}
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(enumerateCmd)
	rootCmd.AddCommand(partitionCmd)
	rootCmd.AddCommand(formatCmd)
	rootCmd.AddCommand(mountCmd)
	rootCmd.AddCommand(umountCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
