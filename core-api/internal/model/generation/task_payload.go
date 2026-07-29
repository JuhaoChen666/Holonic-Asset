package generation

import "encoding/json"

// TaskPayload is owned by generation while the task module treats it as opaque JSON.
type TaskPayload struct {
	ProjectID         uint            `json:"project_id"`
	AssetID           *uint           `json:"asset_id,omitempty"`
	Prompt            string          `json:"prompt"`
	ReferenceMediaIDs []string        `json:"reference_media_ids,omitempty"`
	TargetAssetPaths  []string        `json:"target_asset_paths,omitempty"`
	Parameters        json.RawMessage `json:"parameters,omitempty"`
}

type TaskResult struct {
	AssetID  *uint           `json:"asset_id,omitempty"`
	MediaIDs []string        `json:"media_ids,omitempty"`
	Metadata json.RawMessage `json:"metadata,omitempty"`
}

func NewTaskPayload(request *GenerationRequest) TaskPayload {
	return TaskPayload{
		ProjectID:         request.ProjectID,
		AssetID:           request.AssetID,
		Prompt:            request.Prompt,
		ReferenceMediaIDs: request.ReferenceMediaIDs,
		TargetAssetPaths:  request.TargetAssetPaths,
		Parameters:        request.Parameters,
	}
}
