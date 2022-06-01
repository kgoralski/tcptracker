package servid

import (
	"context"
	"flag"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"tcptracker/cmd/api"
	"tcptracker/internal/connectiontracker"
	"time"

	"github.com/rs/zerolog/log"
)

// App Instance which contains HTTP router
type App struct {
	*http.Server
	mux        *chi.Mux
	handler    *api.Router
	metrics    *prometheus.Registry
	tcpTracker *connectiontracker.Tracker
}

// NewApp creates new App that wraps the dependencies
func NewApp() *App {
	mux := chi.NewRouter()
	metrics := prometheus.NewRegistry()
	params := trackerParams(metrics)
	tracker := connectiontracker.NewTracker(params)
	server := &App{
		mux:        mux,
		handler:    api.NewRouter(mux, metrics),
		metrics:    metrics,
		tcpTracker: tracker,
	}
	server.configureLogger()
	server.routes()
	return server
}

func (app *App) configureLogger() {
	var logJSON bool
	var logLevel int
	flag.BoolVar(&logJSON, "logJSON", false, "configure log format to be PLAIN or JSON")
	flag.IntVar(&logLevel, "logLevel", 0, "configure log level")
	timeFormat := "2006-01-02 15:04:05"
	zerolog.TimeFieldFormat = timeFormat
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
	zerolog.SetGlobalLevel(zerolog.Level(logLevel))
	if !logJSON {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: timeFormat,
		}).With().Timestamp().Logger()
	}
}

func (app *App) routes() {
	app.handler.Routes()
}

// ServerStart launching the HTTP Server
func (app *App) ServerStart() {
	srv := http.Server{Addr: ":8081", Handler: app.mux}
	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())
	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		recSig := <-sig
		log.Info().Msg("Shutting down in progress...")
		// Shutdown signal with grace period of 20 seconds
		shutdownCtx, cancel := context.WithTimeout(serverCtx, 20*time.Second)
		defer cancel()

		go func() {
			<-shutdownCtx.Done()
			if shutdownCtx.Err() == context.DeadlineExceeded {
				log.Fatal().Msg("Graceful shutdown timed out.. forcing exit.")
			}
		}()
		log.Info().Msgf("Graceful shutdown received: %s", recSig.String())
		errClose := app.tcpTracker.Close()
		if errClose != nil {
			log.Err(errClose).Msgf("Closing TCP Tracker with error %s", errClose)
		} else {
			log.Info().Msg("TCP Tracker closed successfully...")
		}
		// Trigger graceful shutdown
		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
		serverStopCtx()
	}()

	// Run the servid
	log.Info().Msg("HTTP server is starting...")
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Info().Msgf("HTTP server ListenAndServe: %v", err)
	}

	// Wait for servid context to be stopped
	<-serverCtx.Done()
}

// TrackHostConnections runs the process of TCP Tracking
func (app *App) TrackHostConnections(ctx context.Context) {
	go app.tcpTracker.Execute(ctx)
}

func trackerParams(metrics *prometheus.Registry) connectiontracker.TrackerParams {
	var deviceName string
	flag.StringVar(&deviceName, "deviceName", "eth0", "Network Interface Device Name to track new connections.")
	flag.Parse()

	firewall, err := connectiontracker.NewFirewall(deviceName)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	params := connectiontracker.TrackerParams{
		DeviceName: deviceName,
		Firewall:   firewall,
		Metrics:    metrics,
	}
	return params
}
