package api

import (
	"context"
	"flag"
	"fmt"
	"gitlab.id.vin/devops/go-kit/src/cmd/config"
	"gitlab.id.vin/devops/go-kit/src/cmd/logger"
	"gitlab.id.vin/devops/go-kit/src/cmd/transport"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	httpPort   = flag.Int("http-port", 8888, "http-port")
	logLevel   = flag.Int("log-level", logger.DebugLevel, "log-level")
	bin        = flag.String("bin", ".", "bin")
	configFile = flag.String("config", "config.yml", "config")
)

func init() {
	//Parse flag
	flag.Parse()
	//Init Logger with log level
	initLogger()
	//Init config
	initConfig()
}

func initLogger() {
	logger.InitLogger(*logLevel)
}

func initConfig() {
	//Init viper to read config
	p, err := filepath.Abs(*bin)
	if err != nil {
		panic(err)
	}
	f := filepath.Join(p, *configFile)
	err = config.InitViper(f)
	if err != nil {
		panic(err)
	}
}

type Server struct {
	Name      string
	HTTPPort  int
	Transport transport.Transport
	OnClose   func()
}

func NewServer(name string, onClose func()) *Server {
	return &Server{
		Name:     name,
		OnClose:  onClose,
		HTTPPort: *httpPort,
	}
}

func (s *Server) AddRouter(transport transport.Transport) {
	s.Transport = transport
}

func (s *Server) Close() {
	s.OnClose()
}

func (s *Server) Run() {
	//Add router
	r := transport.NewRouter(s.Transport)

	//Setup http server
	h := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.HTTPPort),
		Handler: r,
	}

	//Graceful shutdown handle
	idleConnsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)

		<-c

		//Run on close
		s.Close()

		ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

		//A interrupt signal has sent to us, let's shutdown server with gracefully
		logger.Infof("Stopping server...")

		if err := h.Shutdown(ctx); err != nil {
			logger.Errorf("Graceful shutdown has failed with error: %s", err)
		}
		close(idleConnsClosed)
	}()

	go func() {
		logger.Info(nil, fmt.Sprintf("Starting: %v listen server on port %d", s.Name, s.HTTPPort))
		if err := h.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("Run server has failed with error: %s", err)
			//Exit the application if run fail
			os.Exit(1)
		} else {
			logger.Infof("Server was closed by shutdown gracefully")
		}
	}()

	<-idleConnsClosed
}
