package cmd

import (
	"context"
	"database/sql"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/OmAsana/go-yapraktikum-final/migrations"
	"github.com/OmAsana/go-yapraktikum-final/pkg/bonussystem"
	"github.com/OmAsana/go-yapraktikum-final/pkg/logger"
	"github.com/OmAsana/go-yapraktikum-final/pkg/repo"
	"github.com/OmAsana/go-yapraktikum-final/pkg/server"
)

var (
	rootCmd = &cobra.Command{
		PreRunE: setupConfig,
		Run:     run,
	}
)

func Execute() error {
	rootCmd.DisableFlagParsing = true
	return rootCmd.Execute()
}

func rootContext(log *zap.Logger) context.Context {
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		exit := make(chan os.Signal, 1)
		signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
		s := <-exit
		log.Sugar().Infof("Got signal: %d (%s). Shutting down.", s, s.String())
		cancel()
	}()

	return ctx
}

func run(cmd *cobra.Command, args []string) {
	log := logger.NewLogger()
	logger.SetLogLevel(Config.LogLevel)

	defer log.Sync()
	ctx := rootContext(log)

	if err := migrations.ApplyMigrations(Config.DatabaseURI); err != nil {
		log.Sugar().Fatalf("migration: failed to apply migration: %v\n", err)
	}

	db, err := sql.Open("pgx", Config.DatabaseURI)
	if err != nil {
		log.Fatal("could not connect to db", zap.Error(err))
	}

	defer func() {
		_ = db.Close()
	}()

	userRepo, err := repo.UserRepo(db, log)
	if err != nil {
		log.Fatal("could not connect to db", zap.Error(err))
	}

	orderRepo, err := repo.OrderRepo(db, log)
	if err != nil {
		log.Fatal("could not connect to db", zap.Error(err))
	}

	handler := server.NewServer(log, userRepo, orderRepo, Config.Salt)
	srv := &http.Server{Addr: Config.RunAddress, Handler: handler,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		}}

	bonusSystem := bonussystem.NewBonusSystem(Config.AccrualSystemAddress, orderRepo, log)
	g, gCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return bonusSystem.Run(gCtx)
	})
	g.Go(func() error {
		log.Info("Serving", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})
	g.Go(func() error {
		<-gCtx.Done()
		return srv.Shutdown(context.Background())
	})
	if err := g.Wait(); err != nil {
		log.Fatal(err.Error())
	}

}
