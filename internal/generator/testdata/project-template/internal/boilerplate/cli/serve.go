package cli

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"__MODULE_PATH__/internal/boilerplate/app"
	transport "__MODULE_PATH__/internal/boilerplate/http"

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
			httpHandler, err := server.Handler()
			if err != nil {
				return err
			}
			httpServer := server.HTTPServer(httpHandler)
			grpcServer := server.GRPCServer()

			grpcListener, err := net.Listen("tcp", runtime.Config.GRPCAddress())
			if err != nil {
				return err
			}
			defer func() {
				_ = grpcListener.Close()
			}()

			runtime.Logger.Info(
				"starting boilerplate servers",
				"http_addr", runtime.Config.HTTPAddress(),
				"grpc_addr", runtime.Config.GRPCAddress(),
			)

			errCh := make(chan error, 2)
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					errCh <- err
				}
			}()
			go func() {
				if err := grpcServer.Serve(grpcListener); err != nil {
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
				stopDone := make(chan struct{})
				go func() {
					grpcServer.GracefulStop()
					close(stopDone)
				}()
				select {
				case <-stopDone:
				case <-time.After(runtime.Config.ShutdownTimeout):
					grpcServer.Stop()
				}
				return nil
			}
		},
	}
}
