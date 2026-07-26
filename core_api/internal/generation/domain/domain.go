package domain

import "encoding/json"

type RunID uint
type PlanID uint
type StepID uint
type CandidateID uint

// GenerationRequest captures the business intent accepted by generation.
// Kind-specific parameters remain bounded data interpreted by generation.
type GenerationRequest struct {
	ProjectID              uint
	AssetID                uint
	Kind                   RequestKind
	Instruction            string
	ReferenceMediaIDs      []string
	TargetAssetResourceIDs []uint
	Parameters             json.RawMessage
}

type Failure struct {
	Code      string
	Message   string
	Retryable bool
}

type GenerationRun struct {
	ID        RunID
	ProjectID uint
	Request   GenerationRequest
	Status    RunStatus
	PlanID    PlanID
	Failure   *Failure
}

type Plan struct {
	ID    PlanID
	RunID RunID
	Steps []Step
}

type Step struct {
	ID           StepID
	RunID        RunID
	PlanID       PlanID
	Key          string
	Type         string
	Executor     StepExecutor
	Dependencies []StepID
	Parameters   json.RawMessage
	Status       StepStatus
	Attempts     uint
	MaxAttempts  uint
	Failure      *Failure
}

type StepResult struct {
	MediaIDs []string
	Metadata json.RawMessage
}

type Candidate struct {
	ID       CandidateID
	RunID    RunID
	MediaIDs []string
	Status   CandidateStatus
}
