package dto

// CreateProjectPreviewUploadRequest describes a Project preview uploaded directly to object storage.
type CreateProjectPreviewUploadRequest struct {
	ContentType   string `json:"contentType"`
	ContentLength int64  `json:"contentLength"`
}

// ProjectPreviewUploadTarget is the HTTP response for a temporary upload target.
type ProjectPreviewUploadTarget struct {
	ObjectKey   string `json:"objectKey"`
	ObjectURL   string `json:"objectURL"`
	UploadURL   string `json:"uploadURL"`
	UploadToken string `json:"uploadToken"`
}
