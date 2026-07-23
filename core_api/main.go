// Main entry point for the application
package main

import (
	"github.com/1024XEngineer/Holonic-Asset/internal"
	aihandler "github.com/1024XEngineer/Holonic-Asset/internal/ai/handler"
	aiservice "github.com/1024XEngineer/Holonic-Asset/internal/ai/service"
	projecthandler "github.com/1024XEngineer/Holonic-Asset/internal/project/handler"
	projectrepository "github.com/1024XEngineer/Holonic-Asset/internal/project/repository"
	projectdao "github.com/1024XEngineer/Holonic-Asset/internal/project/repository/dao"
	projectservice "github.com/1024XEngineer/Holonic-Asset/internal/project/service"
	taxonomyhandler "github.com/1024XEngineer/Holonic-Asset/internal/taxonomy/handler"
	taxonomyservice "github.com/1024XEngineer/Holonic-Asset/internal/taxonomy/service"
)

func main() {
	projectDao := projectdao.NewMemoryProjectDao()
	projectRepository := projectrepository.NewProjectRepository(projectDao)
	projectService := projectservice.NewProjectService(projectRepository)
	projectHandler := projecthandler.NewProjectHandler(projectService)

	aiService := aiservice.NewAIService()
	aiHandler := aihandler.NewAIHandler(aiService)

	taxonomyService := taxonomyservice.NewAssetDiscoveryService()
	taxonomyHandler := taxonomyhandler.NewTaxonomyHandler(taxonomyService)

	e := internal.Register(nil, projectHandler, aiHandler, taxonomyHandler)
	e.Logger.Fatal(e.Start(":8080"))
}
