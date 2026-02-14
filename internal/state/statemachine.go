package state

import (
	"fmt"
)

// TaskStatus represents the possible states of a task
type TaskStatus string

const (
	TaskStatusOpen       TaskStatus = "open"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
	TaskStatusCancelled  TaskStatus = "cancelled"
)

// Valid state transitions:
// open -> in_progress
// open -> cancelled
// in_progress -> done
// in_progress -> open (reopened)
// done -> open (reopened)
// cancelled -> open (reopened)

var validTransitions = map[TaskStatus][]TaskStatus{
	TaskStatusOpen: {
		TaskStatusInProgress,
		TaskStatusCancelled,
	},
	TaskStatusInProgress: {
		TaskStatusDone,
		TaskStatusOpen,
	},
	TaskStatusDone: {
		TaskStatusOpen,
	},
	TaskStatusCancelled: {
		TaskStatusOpen,
	},
}

// ErrInvalidTransition is returned when a state transition is not allowed
type ErrInvalidTransition struct {
	From     TaskStatus
	To       TaskStatus
	Allowed  []TaskStatus
}

func (e *ErrInvalidTransition) Error() string {
	return fmt.Sprintf("invalid state transition from %s to %s (allowed: %v)", e.From, e.To, e.Allowed)
}

// StateMachine manages task state transitions
type StateMachine struct{}

// NewStateMachine creates a new StateMachine
func NewStateMachine() *StateMachine {
	return &StateMachine{}
}

// CanTransition checks if a transition from current to next state is valid
func (sm *StateMachine) CanTransition(current, next TaskStatus) bool {
	allowed, ok := validTransitions[current]
	if !ok {
		return false
	}

	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

// Transition performs a state transition from current to next state
// Returns an error if the transition is not valid
func (sm *StateMachine) Transition(current, next TaskStatus) error {
	if !sm.CanTransition(current, next) {
		allowed, _ := validTransitions[current]
		return &ErrInvalidTransition{
			From:    current,
			To:      next,
			Allowed: allowed,
		}
	}
	return nil
}

// GetAllowedTransitions returns all possible next states from the current state
func (sm *StateMachine) GetAllowedTransitions(current TaskStatus) []TaskStatus {
	allowed, ok := validTransitions[current]
	if !ok {
		return []TaskStatus{}
	}
	return allowed
}

// IsFinalState returns true if the state is a final state (done or cancelled)
func (sm *StateMachine) IsFinalState(status TaskStatus) bool {
	return status == TaskStatusDone || status == TaskStatusCancelled
}

// IsActiveState returns true if the state is an active state (open or in_progress)
func (sm *StateMachine) IsActiveState(status TaskStatus) bool {
	return status == TaskStatusOpen || status == TaskStatusInProgress
}
