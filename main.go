package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
	"github.com/roysitumorang/bible/config"
	"github.com/roysitumorang/bible/helper"
	"github.com/roysitumorang/bible/router"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func main() {
	ctxt := "Main"
	ctx := context.Background()
	helper.InitLogger()
	cmdVersion := &cobra.Command{
		Use:   "version",
		Short: "print version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("Version: %s\nCommit: %s\nBuild: %s\n", config.Version, config.Commit, config.Build)
		},
	}
	cmdRun := &cobra.Command{
		Use:   "run",
		Short: "run app",
		Run: func(_ *cobra.Command, _ []string) {
			if err := godotenv.Load(".env"); err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrLoad")
				return
			}
			if err := helper.InitHelper(); err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrInitHelper")
				return
			}
			var g errgroup.Group
			service, err := router.MakeHandler(ctx)
			if err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrMakeHandler")
				return
			}
			if err := service.Migration.Migrate(ctx); err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrMigrate")
				return
			}
			g.Go(func() error {
				return service.HTTPServerMain(ctx)
			})
			g.Go(func() error {
				c := cron.New(cron.WithChain(
					cron.Recover(cron.DefaultLogger),
				))
				c.Start()
				helper.Log(ctx, zap.InfoLevel, "cron: scheduled tasks running!...", ctxt, "")
				return nil
			})
			if err := g.Wait(); err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrWait")
			}
		},
	}
	cmdMigration := &cobra.Command{
		Use:   "migration",
		Short: "new/run migration",
		Args: func(_ *cobra.Command, args []string) (err error) {
			if len(args) == 0 {
				err = errors.New("requires at least 1 arg (new|run")
				return
			}
			if args[0] != "new" && args[0] != "run" {
				err = fmt.Errorf("invalid first flag specified: %s", args[0])
			}
			return
		},
		Run: func(_ *cobra.Command, args []string) {
			now := time.Now()
			if err := godotenv.Load(".env"); err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrLoad")
				return
			}
			if err := helper.InitHelper(); err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrInitHelper")
				return
			}
			service, err := router.MakeHandler(ctx)
			if err != nil {
				helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrMakeHandler")
				return
			}
			switch args[0] {
			case "new":
				if err := service.Migration.CreateMigrationFile(ctx); err != nil {
					helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrCreateMigrationFile")
					return
				}
				duration := time.Since(now)
				helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("creating migration successfully in %s", duration.String()), ctxt, "")
			case "run":
				if err := service.Migration.Migrate(ctx); err != nil {
					helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrMigrate")
					return
				}
				duration := time.Since(now)
				helper.Log(ctx, zap.InfoLevel, fmt.Sprintf("running migration successfully in %s", duration.String()), ctxt, "")
			}
		},
	}
	rootCmd := &cobra.Command{Use: config.AppName}
	rootCmd.AddCommand(
		cmdVersion,
		cmdRun,
		cmdMigration,
	)
	rootCmd.SuggestionsMinimumDistance = 1
	if err := rootCmd.Execute(); err != nil {
		helper.Capture(ctx, zap.ErrorLevel, err, ctxt, "ErrExecute")
	}
}
