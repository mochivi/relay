package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mochivi/relay/internal/config"
	"github.com/mochivi/relay/internal/proxy"
	"github.com/mochivi/relay/internal/router"
	"github.com/mochivi/relay/internal/service"
)

func main() {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/example.yaml"
	}
	cfgFile, err := os.OpenFile(configPath, os.O_RDONLY, 0444)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}
	cfg, err := config.ParseConfig(cfgFile)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}

	services := make(map[string]*service.Service, 0)
	for _, serviceConfig := range cfg.Services {
		service, err := service.NewService(*serviceConfig)
		if err != nil {
			log.Fatalf("Failed to parse service: %v", err)
		}
		services[service.Name] = service
	}

	router, err := router.NewRouter(cfg.Routes, services)
	if err != nil {
		log.Fatalf("Failed to create router: %v", err)
	}

	proxy := proxy.NewProxy(*cfg.Global, router)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := proxy.Start(); err != nil {
			log.Fatalf("Proxy shutdown: %v", err)
		}
	}()

	<-sigs

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := proxy.Shutdown(ctx); err != nil {
		log.Fatalf("Forced shutdown")
	} else {
		fmt.Println("Server exited gracefully")
	}
}
