package main

import (
	"log"

	"github.com/Zhaoyikaiii/clawteam/internal/agent"
	"github.com/Zhaoyikaiii/clawteam/internal/config"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	log.Printf("Starting Agent Runtime on %s", cfg.GetAPIAddr())
	log.Printf("Using LLM model: %s", cfg.LLM.DefaultModel)

	// Create runtime
	runtime := agent.NewRuntime()

	// Register mock agent
	runtime.RegisterAgent(&agent.Agent{
		ID:           "summary-agent",
		Name:         "SummaryAgent",
		Model:        cfg.LLM.DefaultModel,
		SystemPrompt: "You are a helpful assistant that summarizes discussions.",
	})

	log.Println("Agent Runtime started successfully")
}
