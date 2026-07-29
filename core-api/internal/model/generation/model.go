package generation

import (
	"encoding/json"

	taskdomain "github.com/1024XEngineer/Holonic-Asset/internal/module/task"
)

type RunID uint
type PlanID uint
type StepID uint
type RunListStatus string

// GenerationRequest captures the business intent accepted by generation.
// Kind-specific parameters remain bounded data interpreted by generation.
type GenerationRequest struct {
	ProjectID         uint
	AssetID           uint
	Kind              RequestKind
	Prompt            string
	ReferenceMediaIDs []string
	TargetAssetPaths  []string
	Parameters        json.RawMessage
}

type Failure struct {
	Code    string
	Message string
}

type GenerationRun struct {
	ID             RunID
	PlanningTaskID *uint
	ProjectID      uint
	Request        GenerationRequest
	PlanID         PlanID
	Lifecycle      RunLifecycle
	Failure        *Failure
}

// GenerationDetail is assembled from generation records and task state.
type GenerationDetail struct {
	Run   GenerationRun
	Steps []StepDetail
}

// RunListQuery contains the client-facing list semantics. The service expands
// status and asset scope into repository filters.
type RunListQuery struct {
	ProjectID uint
	AssetID   *uint
	Status    RunListStatus
	Limit     int
	Cursor    string
}

type RunListFilter struct {
	ProjectID        uint
	AssetID          *uint
	Lifecycles       []RunLifecycle
	IncludeTaskTypes []TaskType
	ExcludeTaskTypes []TaskType
	Limit            int
	Cursor           string
}

type RunListPage struct {
	Runs       []GenerationRun
	NextCursor string
}

// StepDetail carries task-owned execution state without persisting a duplicate
// status in the generation module.
type StepDetail struct {
	Step       Step
	TaskStatus *taskdomain.Status
}

type Plan struct {
	ID    PlanID
	RunID RunID
	Steps []Step
}

type Step struct {
	ID           StepID
	TaskID       *uint
	RunID        RunID
	PlanID       PlanID
	Key          string
	Type         string
	Executor     StepExecutor
	Dependencies []StepID
	Parameters   json.RawMessage
}

type StepResult struct {
	MediaIDs []string
	Metadata json.RawMessage
}
