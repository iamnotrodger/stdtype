package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/iamnotrodger/stdytype/internal/local"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ContextWithSignal(sigs ...os.Signal) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)

	go func() {
		<-ch
		cancel()
	}()

	return ctx
}

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "stdtype",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := viper.New()

			if err := v.BindPFlags(cmd.Flags()); err != nil {
				return errors.Wrap(err, "viper.BindPFlags")
			}

			v.SetEnvPrefix("stdtype")
			v.AutomaticEnv()

			configDir, err := local.ConfigDir()
			if err != nil {
				return errors.Wrap(err, "local.ConfigDir")
			}

			settingsPath := filepath.Join(configDir, "settings.json")
			settingsRepo, err := local.NewSettingsRepo(v, settingsPath)
			if err != nil {
				return errors.Wrap(err, "local.NewSettingsRepo")
			}

			settings, err := settingsRepo.GetSettings()
			if err != nil {
				return errors.Wrap(err, "settingsRepo.GetSettings")
			}

			fmt.Printf("%+v", settings)

			return nil
		},
	}

	SetSettingsFlag(cmd.Flags())

	return cmd
}

func main() {
	ctx := ContextWithSignal(syscall.SIGTERM, syscall.SIGINT)
	rootCmd := NewRootCmd()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
		os.Exit(1)
	}
}
