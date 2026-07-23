package dto

// CreateMediaUploadRequest identifies the asset resource that needs a temporary upload target.
type CreateMediaUploadRequest struct {
	AssetID         uint   `json:"assetId"`
	AssetResourceID uint   `json:"assetResourceId"`
	ContentType     string `json:"contentType"`
}

// ObjectUploadTarget contains the stable object reference and temporary upload URL.
type ObjectUploadTarget struct {
	ObjectKey string `json:"objectKey"`
	UploadURL string `json:"uploadUrl"`
}
