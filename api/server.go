package api

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/1infras/go-kit/config"
	"github.com/1infras/go-kit/driver/consul"
	"github.com/1infras/go-kit/logger"
	"github.com/1infras/go-kit/transport"
)

var (
	httpPort      = flag.Int("http-port", 8888, "Listen port of HTTP Server")
	logLevel      = flag.String("log-level", "debug", "Lowest level of logger")
	configNames   = flag.String("config-name", "config.yaml", "The name of local config files or remote keys or environment, separate with comma (,)")
	configType    = flag.String("config-type", "skip", "Choose one of config types to use, local | remote | skip")
	prefixEnv     = flag.String("prefix-env", "", "Set prefix environment to read config from environment")
	connectConsul = flag.Bool("connect-consul", false, "Auto connect to consul by auto read configuration from environment")
)

func init() {
	// Parse flag
	flag.Parse()
	// Init Logger with log level
	initLogger()
	// Init Config
	initConfig()
}

func initLogger() {
	logger.InitLogger(logger.LogLevel(*logLevel))
}

func initConfig() {
	if *configType != "skip" {
		cfg := &config.Config{
			ConfigType:        *configType,
			PrefixEnvironment: *prefixEnv,
			Names:             *configNames,
			ConsulKV:          nil,
		}
		if *configType == "remote" && *connectConsul {
			kv, err := consul.NewConsul(nil)
			if err != nil {
				panic(fmt.Errorf("auto connect consul have error: %s", err.Error()))
			}
			cfg.ConsulKV = kv
		}
		err := config.AutomateReadConfig(cfg)
		if err != nil {
			panic(fmt.Errorf("auto read config have error: %s", err.Error()))
		}
	}
}

// Server - HTTP Server
type httpServer struct {
	name        string
	httpPort    int
	pathPrefix  string
	strictSlash bool
	routes      []*transport.Route
	readConfig  *config.Config
	onClose     func()
}

// NewServer - New a HTTP Server with name and close function when it's close
func NewServer(name string, onClose func()) *httpServer {
	return &httpServer{
		name:     name,
		httpPort: *httpPort,
		onClose:  onClose,
	}
}

// AddRouter - Set a router for HTTP Server
func (_this *httpServer) AddRouter(pathPrefix string, strictSlash bool, routes []*transport.Route) {
	_this.pathPrefix = pathPrefix
	_this.strictSlash = strictSlash
	_this.routes = routes
}

// AddReadConfig -
func (_this *httpServer) AddReadConfig(cfg *config.Config) {
	_this.readConfig = cfg
}

// Run - Listen and Serve HTTP Server
func (_this *httpServer) Run() {
	if _this.readConfig != nil {
		err := config.AutomateReadConfig(_this.readConfig)
		if err != nil {
			panic(fmt.Errorf("auto read config have error: %s", err.Error()))
		}
	}

	// Add router
	r := transport.NewRouter(_this.pathPrefix, _this.strictSlash, _this.routes)

	// Setup http server
	h := &http.Server{
		Addr:    fmt.Sprintf(":%d", _this.httpPort),
		Handler: r,
	}

	// Graceful shutdown handle
	idleConnectionsClosed := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)
		signal.Notify(c, syscall.SIGTERM)

		<-c

		// Run on close
		_this.onClose()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		// A interrupt signal has sent to us, let's shutdown server with gracefully
		logger.Info("Stopping server...")

		if err := h.Shutdown(ctx); err != nil {
			logger.Errorf("Graceful shutdown has failed with error: %s", err)
		}
		close(idleConnectionsClosed)
	}()

	go func() {
		logger.Infof("Starting: %v listen server on port %v", _this.name, _this.httpPort)
		if err := h.ListenAndServe(); err != http.ErrServerClosed {
			logger.Errorf("Run server has failed with error: %s", err)
			// Exit the application if run fail
			os.Exit(1)
		} else {
			logger.Infof("Server was closed by shutdown gracefully")
		}
	}()

	<-idleConnectionsClosed
}
