package main

import (
	"github.com/labstack/echo/v4"

	"github.com/1024XEngineer/Holonic-Asset/internal/handler"
	"github.com/1024XEngineer/Holonic-Asset/internal/repository"
	"github.com/1024XEngineer/Holonic-Asset/internal/repository/dao"
	"github.com/1024XEngineer/Holonic-Asset/internal/router"
	appservice "github.com/1024XEngineer/Holonic-Asset/internal/service"
)

type App struct {
	engine *echo.Echo
}

func InitServer() *App {
	projectDao := dao.NewMemoryProjectDao()
	projectRepository := repository.NewProjectRepository(projectDao)
	projectService := appservice.NewProjectService(projectRepository)
	projectHandler := handler.NewProjectHandler(projectService)

	generationService := appservice.NewGenerationService(nil, nil, nil)
	generationHandler := handler.NewGenerationHandler(generationService)

	mediaService := appservice.NewMediaService()
	mediaHandler := handler.NewMediaHandler(mediaService)

	return &App{
		engine: router.Register(nil, projectHandler, generationHandler, mediaHandler),
	}
}

func (a *App) Start(address string) error {
	return a.engine.Start(address)
}
