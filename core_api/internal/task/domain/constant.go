package domain

type Status uint
type TaskType string
type RunStatus string
type StepStatus string
type CandidateStatus string

const (
	StatusPending Status = iota
	StatusProcessing
	StatusCompleted
	StatusFailed
	StatusCancelled
)

const (
	RunStatusPending             RunStatus = "pending"
	RunStatusPlanning            RunStatus = "planning"
	RunStatusPlanned             RunStatus = "planned"
	RunStatusRunning             RunStatus = "running"
	RunStatusPostProcessing      RunStatus = "post_processing"
	RunStatusWaitingConfirmation RunStatus = "waiting_confirmation"
	RunStatusCompleted           RunStatus = "completed"
	RunStatusFailed              RunStatus = "failed"
	RunStatusCancelled           RunStatus = "cancelled"
)

const (
	StepStatusPending   StepStatus = "pending"
	StepStatusReady     StepStatus = "ready"
	StepStatusRunning   StepStatus = "running"
	StepStatusSucceeded StepStatus = "succeeded"
	StepStatusFailed    StepStatus = "failed"
	StepStatusRetryWait StepStatus = "retry_wait"
	StepStatusCancelled StepStatus = "cancelled"
	StepStatusSkipped   StepStatus = "skipped"
)

const (
	CandidateStatusPending   CandidateStatus = "pending"
	CandidateStatusConfirmed CandidateStatus = "confirmed"
	CandidateStatusRejected  CandidateStatus = "rejected"
)

const (
	GenerateCharacterProtoType   TaskType = "generateCharacterProtoType"
	GenerateCharacterAnimation   TaskType = "generateCharacterAnimation"
	RegenerateCharacterProtoType TaskType = "regenerateCharacterProtoType"
	RegenerateCharacterAnimation TaskType = "regenerateCharacterAnimation"
	RegenerateCharacterFrames    TaskType = "regenerateCharacterFrames"

	GenerateObjectProtoType   TaskType = "generateObjectProtoType"
	GenerateObjectAnimation   TaskType = "generateObjectAnimation"
	RegenerateObjectProtoType TaskType = "regenerateObjectProtoType"
	RegenerateObjectAnimation TaskType = "regenerateObjectAnimation"
	RegenerateObjectFrames    TaskType = "regenerateObjectFrames"

	GenerateTileSet TaskType = "generateTileSet"
	RegenerateItem  TaskType = "regenerateItem"
	RegenerateTiles TaskType = "regenerateTiles"
)

type OutboxStatus uint

const (
	OutboxPending   OutboxStatus = 0 // waiting to be published to River
	OutboxPublished OutboxStatus = 1 // successfully published to River
)
