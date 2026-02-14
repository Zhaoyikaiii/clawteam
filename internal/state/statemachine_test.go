package state

import (
	"testing"
)

func TestStateMachine_CanTransition(t *testing.T) {
	sm := NewStateMachine()

	tests := []struct {
		name     string
		current  TaskStatus
		next     TaskStatus
		expected bool
	}{
		{"open to in_progress", TaskStatusOpen, TaskStatusInProgress, true},
		{"open to cancelled", TaskStatusOpen, TaskStatusCancelled, true},
		{"open to done", TaskStatusOpen, TaskStatusDone, false},
		{"in_progress to done", TaskStatusInProgress, TaskStatusDone, true},
		{"in_progress to open", TaskStatusInProgress, TaskStatusOpen, true},
		{"in_progress to cancelled", TaskStatusInProgress, TaskStatusCancelled, false},
		{"done to open", TaskStatusDone, TaskStatusOpen, true},
		{"done to in_progress", TaskStatusDone, TaskStatusInProgress, false},
		{"cancelled to open", TaskStatusCancelled, TaskStatusOpen, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sm.CanTransition(tt.current, tt.next); got != tt.expected {
				t.Errorf("CanTransition() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStateMachine_Transition(t *testing.T) {
	sm := NewStateMachine()

	tests := []struct {
		name    string
		current TaskStatus
		next    TaskStatus
		wantErr bool
	}{
		{"valid transition", TaskStatusOpen, TaskStatusInProgress, false},
		{"invalid transition", TaskStatusOpen, TaskStatusDone, true},
		{"reopen done task", TaskStatusDone, TaskStatusOpen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.Transition(tt.current, tt.next)
			if (err != nil) != tt.wantErr {
				t.Errorf("Transition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStateMachine_GetAllowedTransitions(t *testing.T) {
	sm := NewStateMachine()

	tests := []struct {
		name         string
		current      TaskStatus
		minExpected  int
	}{
		{"open states", TaskStatusOpen, 2},
		{"in_progress states", TaskStatusInProgress, 2},
		{"done states", TaskStatusDone, 1},
		{"cancelled states", TaskStatusCancelled, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			states := sm.GetAllowedTransitions(tt.current)
			if len(states) < tt.minExpected {
				t.Errorf("GetAllowedTransitions() = %v, want at least %d states", states, tt.minExpected)
			}
		})
	}
}

func TestStateMachine_IsFinalState(t *testing.T) {
	sm := NewStateMachine()

	if !sm.IsFinalState(TaskStatusDone) {
		t.Error("Done should be a final state")
	}
	if !sm.IsFinalState(TaskStatusCancelled) {
		t.Error("Cancelled should be a final state")
	}
	if sm.IsFinalState(TaskStatusOpen) {
		t.Error("Open should not be a final state")
	}
	if sm.IsFinalState(TaskStatusInProgress) {
		t.Error("InProgress should not be a final state")
	}
}

func TestStateMachine_IsActiveState(t *testing.T) {
	sm := NewStateMachine()

	if !sm.IsActiveState(TaskStatusOpen) {
		t.Error("Open should be an active state")
	}
	if !sm.IsActiveState(TaskStatusInProgress) {
		t.Error("InProgress should be an active state")
	}
	if sm.IsActiveState(TaskStatusDone) {
		t.Error("Done should not be an active state")
	}
	if sm.IsActiveState(TaskStatusCancelled) {
		t.Error("Cancelled should not be an active state")
	}
}
