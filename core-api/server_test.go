package main

import (
	"testing"

	"github.com/1024XEngineer/Holonic-Asset/internal/repository/dao"
)

type projectDaoStub struct {
	dao.ProjectDao
}

func TestNewAppBuildsApplication(t *testing.T) {
	app := NewApp(&projectDaoStub{})
	if app == nil {
		t.Fatal("expected server application")
	}
	if app.engine == nil {
		t.Fatal("expected server engine")
	}
}
