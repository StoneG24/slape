/*
Package SLaPE is a binary that orchestrates containers using docker on the local computer.

Usage:

	./slape

Containerized models are spawned as needed adhering to a pipeline system.
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/StoneG24/slape/cmd/api"
	"github.com/StoneG24/slape/cmd/pipeline"
	_ "github.com/StoneG24/slape/docs"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

var (
	s = pipeline.SimplePipeline{
		// updates after created
		Model:          "",
		ContextBox:     pipeline.ContextBox{},
		Tools:          pipeline.Tools{},
		Active:         true,
		ContainerImage: "",
		DockerClient:   nil,
		GPU:            false,
	}

	c = pipeline.ChainofModels{
		// updates after created
		Models:         []string{},
		ContextBox:     pipeline.ContextBox{},
		Tools:          pipeline.Tools{},
		Active:         true,
		ContainerImage: "",
		DockerClient:   nil,
		GPU:            false,
	}

	d = pipeline.DebateofModels{
		// updates after created
		Models:         []string{},
		ContextBox:     pipeline.ContextBox{},
		Tools:          pipeline.Tools{},
		Active:         true,
		ContainerImage: "",
		DockerClient:   nil,
		GPU:            false,
	}
)

// @title SLaPE API
// @version 1.0
// @description This is a swagger for SLaPE
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:3069
// @BasePath /
func main() {
	app := fiber.New()

	// channel for managing pipelines
	// keystone := make(chan pipeline.Pipeline)
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	// TODO(v) switch to router for method checking and routing
	app.Post("/simple", pipeline.SimplePipelineGenerateRequest)
	// http.HandleFunc("/smplsetup", s.SimplePipelineSetupRequest)
	// http.HandleFunc("/cot", c.ChainPipelineGenerateRequest)
	// http.HandleFunc("/cotsetup", c.ChainPipelineSetupRequest)
	// http.HandleFunc("/debate", d.DebatePipelineGenerateRequest)
	// http.HandleFunc("/debsetup", d.DebatePipelineSetupRequest)
	app.Get("/swagger/*", swagger.HandlerDefault) // default

	//http.HandleFunc("/moe", simplerequest)
	http.HandleFunc("/getmodels", api.GetModels)

	// Listen from a different goroutine
	go func() {
		if err := app.Listen(":8080"); err != nil {
			log.Panic(err)
		}
	}()

	c := make(chan os.Signal, 1)                    // Create channel to signify a signal being sent
	signal.Notify(c, os.Interrupt, syscall.SIGTERM) // When an interrupt or termination signal is sent, notify the channel

	_ = <-c // This blocks the main thread until an interrupt is received
	fmt.Println("Gracefully shutting down...")
	_ = app.Shutdown()

	fmt.Println("Running cleanup tasks...")

	// Your cleanup tasks go here
	// db.Close()
	// redisConn.Close()
	fmt.Println("Fiber was successful shutdown.")
}
