package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

type User struct {
	log *zap.Logger
}

func NewUserHandle(log *zap.Logger) *User {
	return &User{log: log}
}

type Route interface {
	http.Handler

	// Pattern reports the path at which this is registered.
	Pattern() string
}

func (*User) Pattern() string {
	return "/user"
}

// ServeHTTP handles an HTTP request to the /echo endpoint.
func (h *User) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, err := io.Copy(w, r.Body); err != nil {
		h.log.Warn("Failed to handle request", zap.Error(err))
	}
	_, err := io.ReadAll(r.Body)

	if err != nil {
		h.log.Error("Failed to read request", zap.Error(err))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func NewHTTPServer(lc fx.Lifecycle, mux *http.ServeMux) *http.Server {
	srv := &http.Server{Addr: ":8080", Handler: mux}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			fmt.Println("Starting HTTP server at", srv.Addr)
			go srv.Serve(ln)
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
	return srv
}

// NewServeMux builds a ServeMux that will route requests
// to the given routes.
func NewServeMux(route Route) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle(route.Pattern(), route)
	return mux
}

var Module = fx.Module("server", fx.Provide(
	NewHTTPServer,
	NewServeMux,
	fx.Annotate(
		NewUserHandle,
		fx.As(new(Route)),
	),
	zap.NewExample,
),
	fx.Invoke(func(*http.Server) {}),
)

func main() {
	fx.New(Module).Run()
}
