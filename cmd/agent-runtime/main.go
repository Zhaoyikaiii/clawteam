package main

import "github.com/wangxumarshall/clawteam/internal/agent"

func main() {
    runtime := agent.NewRuntime()
    
    runtime.RegisterAgent(&agent.Agent{
        ID:          "summary-agent",
        Name:        "SummaryAgent",
        SystemPrompt: "You are a helpful assistant.",
    })

    println("Agent Runtime started successfully")
}
