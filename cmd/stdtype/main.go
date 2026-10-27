package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/debug"
	"syscall"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/iamnotrodger/stdytype/internal/local"
	"github.com/iamnotrodger/stdytype/internal/words"
	"github.com/iamnotrodger/stdytype/ui"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "dev"

func init() {
	if version != "dev" {
		return
	}

	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		version = info.Main.Version
	}
}

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
		Use:           "stdtype",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
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

			wordSource, err := words.NewWordSource(time.Now().UnixNano())
			if err != nil {
				return errors.Wrap(err, "words.NewWordSource")
			}

			statsRepo, err := local.NewStatsRepo(configDir, "results.json")
			if err != nil {
				return errors.Wrap(err, "local.NewResultRepo")
			}

			settingsPath := filepath.Join(configDir, "settings.json")
			v.SetConfigFile(settingsPath)
			v.SetConfigType("json")

			settingsRepo, err := local.NewSettingsRepo(v, settingsPath)
			if err != nil {
				return errors.Wrap(err, "local.NewSettingsRepo")
			}

			settings, err := settingsRepo.GetSettings()
			if err != nil {
				return errors.Wrap(err, "settingsRepo.GetSettings")
			}

			if err := settings.Validate(wordSource.Lists()); err != nil {
				return errors.Wrap(err, "invalid settings")
			}

			model, err := ui.NewModel(wordSource, statsRepo, settingsRepo)
			if err != nil {
				return errors.Wrap(err, "ui.NewModel")
			}

			program := tea.NewProgram(
				model,
				tea.WithContext(cmd.Context()),
				tea.WithoutSignalHandler(),
			)

			if _, err := program.Run(); err != nil {
				return errors.Wrap(err, "program.Run")
			}

			return nil
		},
	}

	cmd.SetVersionTemplate("{{.Version}}\n")
	SetSettingsFlag(cmd.Flags())

	return cmd
}

func main() {
	ctx := ContextWithSignal(syscall.SIGTERM, syscall.SIGINT)
	rootCmd := NewRootCmd()

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
