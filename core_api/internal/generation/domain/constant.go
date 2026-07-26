package domain

type RunStatus string
type StepStatus string
type CandidateStatus string
type StepExecutor string
type RequestKind string

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
	StepExecutorAI          StepExecutor = "ai"
	StepExecutorAssetWorker StepExecutor = "asset_worker"
)

const (
	RequestKindGenerateCharacter      RequestKind = "generate_character"
	RequestKindGenerateProjectPreview RequestKind = "generate_project_preview"
	RequestKindGenerateTileSetItem    RequestKind = "generate_tileset_item"
	RequestKindEditTileSetItem        RequestKind = "edit_tileset_item"
	RequestKindGenerateObject         RequestKind = "generate_object"
	RequestKindGenerateSceneryLayer   RequestKind = "generate_scenery_layer"
	RequestKindEditSceneryLayer       RequestKind = "edit_scenery_layer"
	RequestKindGenerateAnimation      RequestKind = "generate_animation"
	RequestKindEditFrame              RequestKind = "edit_frame"
	RequestKindGenerateUI             RequestKind = "generate_ui"
	RequestKindEditUIComponent        RequestKind = "edit_ui_component"
	RequestKindGenerateAudio          RequestKind = "generate_audio"
	RequestKindEditAudio              RequestKind = "edit_audio"
)
