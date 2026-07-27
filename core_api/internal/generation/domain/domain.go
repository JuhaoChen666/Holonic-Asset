package domain

import (
	"encoding/json"

	taskdomain "github.com/1024XEngineer/Holonic-Asset/internal/task/domain"
)

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
	Prompt                 string
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
	Status    taskdomain.RunStatus
	PlanID    PlanID
	Failure   *Failure
}

// GenerationDetail is the read model returned for one generation lifecycle.
type GenerationDetail struct {
	Run        GenerationRun
	Steps      []Step
	Candidates []Candidate
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
	Status       taskdomain.StepStatus
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
	Status   taskdomain.CandidateStatus
}

// ConfirmCandidateCommand contains the asset target required to apply an
// accepted candidate without making the asset module query generation state.
type ConfirmCandidateCommand struct {
	RunID                  RunID
	CandidateID            CandidateID
	ProjectID              uint
	AssetID                uint
	Kind                   RequestKind
	TargetAssetResourceIDs []uint
	MediaIDs               []string
}
