package dto_test

import (
	"encoding/json"
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/media/dto"
)

func TestCreateMediaUploadRequestJSONContract(t *testing.T) {
	request := dto.CreateMediaUploadRequest{
		AssetID:         42,
		AssetResourceID: 99,
		ContentType:     "image/png",
	}

	encoded, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	want := `{"assetId":42,"assetResourceId":99,"contentType":"image/png"}`
	if string(encoded) != want {
		t.Fatalf("expected %s, got %s", want, encoded)
	}
}

func TestObjectUploadTargetJSONContract(t *testing.T) {
	target := dto.ObjectUploadTarget{
		ObjectKey: "assets/42/resources/99/source.png",
		UploadURL: "https://storage.example/upload",
	}

	encoded, err := json.Marshal(target)
	if err != nil {
		t.Fatalf("marshal target: %v", err)
	}

	want := `{"objectKey":"assets/42/resources/99/source.png","uploadUrl":"https://storage.example/upload"}`
	if string(encoded) != want {
		t.Fatalf("expected %s, got %s", want, encoded)
	}
}
