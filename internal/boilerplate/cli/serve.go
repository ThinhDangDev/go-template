package cli

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ThinhDangDev/go-template/internal/boilerplate/app"
	transport "github.com/ThinhDangDev/go-template/internal/boilerplate/http"

	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the boilerplate HTTP API",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			runtime, err := app.Bootstrap(ctx)
			if err != nil {
				return err
			}
			defer func() {
				shutdownCtx, cancel := context.WithTimeout(context.Background(), runtime.Config.ShutdownTimeout)
				defer cancel()
				_ = runtime.Close(shutdownCtx)
			}()

			server := transport.NewServer(runtime)
			httpServer := server.HTTPServer(server.Handler())

			runtime.Logger.Info("starting boilerplate server", "addr", runtime.Config.HTTPAddress())

			errCh := make(chan error, 1)
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					errCh <- err
				}
			}()

			signalCh := make(chan os.Signal, 1)
			signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
			defer signal.Stop(signalCh)

			select {
			case err := <-errCh:
				return err
			case sig := <-signalCh:
				runtime.Logger.Info("received shutdown signal", "signal", sig.String())
				shutdownCtx, cancel := context.WithTimeout(context.Background(), runtime.Config.ShutdownTimeout)
				defer cancel()

				if err := httpServer.Shutdown(shutdownCtx); err != nil {
					return err
				}
				return nil
			}
		},
	}
}
