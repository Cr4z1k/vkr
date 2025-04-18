package main

import (
	"log"

	"github.com/Cr4z1k/vkr/internal/docker_cli"
	cfg_service "github.com/Cr4z1k/vkr/internal/service/parser"
	"github.com/Cr4z1k/vkr/internal/transport/rest"
	"github.com/Cr4z1k/vkr/internal/transport/rest/handlers"
	"github.com/Cr4z1k/vkr/internal/transport/rest/handlers/configs"
	"github.com/docker/docker/client"
)

func main() {
	s := new(rest.Server)

	// init tools
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatalf("failed to create Docker client: %v", err)
	}

	docker := docker_cli.New(dockerCli)

	// init services
	parserService := cfg_service.New()

	// init handlers
	cfgHandler := configs.New(parserService, docker)

	// main handler init
	mainHanlder := handlers.New(cfgHandler)

	if err := s.Run("8080", mainHanlder.InitRoutes()); err != nil {
		log.Fatal("Fatal start server: ", err.Error())
	}
}
