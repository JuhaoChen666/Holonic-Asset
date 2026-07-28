package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/dto"
)

func TestCreateProjectPreviewUploadRequestJSONContract(t *testing.T) {
	request := dto.CreateProjectPreviewUploadRequest{
		ContentType:   "image/png",
		ContentLength: 8,
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	want := `{"contentType":"image/png","contentLength":8}`
	if string(encoded) != want {
		t.Fatalf("expected %s, got %s", want, encoded)
	}
}

func TestProjectPreviewUploadTargetJSONContract(t *testing.T) {
	target := dto.ProjectPreviewUploadTarget{
		ObjectKey:   "users/7/project-previews/uuid",
		ObjectURL:   "https://media.example.com/users/7/project-previews/uuid",
		UploadURL:   "https://upload.qiniup.com",
		UploadToken: "access-key:signature:policy",
	}

	encoded, err := json.Marshal(target)
	if err != nil {
		t.Fatalf("marshal target: %v", err)
	}

	want := `{"objectKey":"users/7/project-previews/uuid","objectURL":"https://media.example.com/users/7/project-previews/uuid","uploadURL":"https://upload.qiniup.com","uploadToken":"access-key:signature:policy"}`
	if string(encoded) != want {
		t.Fatalf("expected %s, got %s", want, encoded)
	}
}
