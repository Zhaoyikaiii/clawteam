// Package agent provides the core agent execution logic
package agent

import (
    "context"
    "fmt"
    "log/slog"
)

// Agent represents an AI agent that can process requests
type Agent struct {
    ID       string
    Name     string
    Model     string
    SystemPrompt string
    logger    *slog.Logger
}

// Runtime manages agent execution
type Runtime struct {
    agents map[string]*Agent
    logger  *slog.Logger
}

// NewRuntime creates a new agent runtime
func NewRuntime() *Runtime {
    return &Runtime{
        agents: make(map[string]*Agent),
    }
}

// RegisterAgent adds a new agent to the runtime
func (r *Runtime) RegisterAgent(agent *Agent) {
    r.agents[agent.ID] = agent
}

// ProcessRequest processes a request from Social Brain
func (r *Runtime) ProcessRequest(ctx context.Context, req *Request) (*Response, error) {
    agent, ok := r.agents[req.AgentID]
    if !ok {
        return nil, fmt.Errorf("agent not found: %s", req.AgentID)
    }

    r.logger.Info("processing request for agent", "agent_id", req.AgentID)

    // TODO: Call LLM with agent's system prompt
    // For now, return a mock response
    return &Response{
        Content: fmt.Sprintf("Agent %s processed: %s", agent.Name, req.Input),
        AgentID: agent.ID,
    }, nil
}

// Request represents an incoming request
type Request struct {
    AgentID string
    Input    string
    Context  map[string]interface{}
}

// Response represents the agent's response
type Response struct {
    Content string
    AgentID string
}
