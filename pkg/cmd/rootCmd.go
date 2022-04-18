package cmd

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/OmAsana/go-yapraktikum-final/pkg/logger"
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

	handler := server.NewServer(log)
	srv := &http.Server{Addr: Config.RunAddress, Handler: handler,
		BaseContext: func(listener net.Listener) context.Context {
			return ctx
		}}

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Sugar().Infof("Serving on: %s", srv.Addr)
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
