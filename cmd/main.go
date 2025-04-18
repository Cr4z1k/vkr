package main

import (
	"log"

	cfg_service "github.com/Cr4z1k/vkr/internal/service/parser"
	"github.com/Cr4z1k/vkr/internal/transport/rest"
	"github.com/Cr4z1k/vkr/internal/transport/rest/handlers"
	"github.com/Cr4z1k/vkr/internal/transport/rest/handlers/configs"
)

func main() {
	s := new(rest.Server)

	// init services
	parserService := cfg_service.New()

	// init handlers
	cfgHandler := configs.New(parserService)

	// main handler init
	mainHanlder := handlers.New(cfgHandler)

	if err := s.Run("8080", mainHanlder.InitRoutes()); err != nil {
		log.Fatal("Fatal start server: ", err.Error())
	}
}
