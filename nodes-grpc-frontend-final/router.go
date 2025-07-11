package main

import (
	"embed"
	"html/template"
	"net/http"
	"nodes-grpc-fe/handlers"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

//go:embed assets/* views/*
var contents embed.FS

func setRouters(app *gin.Engine, resty *resty.Client) {
	templ := template.Must(template.New("").ParseFS(contents, "views/*.html"))
	app.SetHTMLTemplate(templ)

	app.StaticFS("/public", http.FS(contents))

	apiClient := handlers.NewApiClient(resty)

	app.GET("/", handlers.RedirectToLogin())

	// login
	app.GET("/login", handlers.LoginPageHandlers())
	app.POST("/login", handlers.LoginHandlers(apiClient))

	// dashboard
	app.GET("/dashboard", handlers.AddTokenHeader(), handlers.DashboardPageHandlers(apiClient))

	// create cluster
	app.GET("/create-cluster", handlers.AddTokenHeader(), handlers.CreateClusterPageHandler(apiClient))
	app.POST("/create-cluster", handlers.AddTokenHeader(), handlers.CreateClusterHandler(apiClient))

	// cluster check
	app.GET("/cluster/:cluster_id", handlers.AddTokenHeader(), handlers.AccessCluster(apiClient))
	app.DELETE("/cluster/:cluster_id/delete", handlers.AddTokenHeader(), handlers.DeleteCluster(apiClient))
	app.GET("/cluster/:cluster_id/status", handlers.AddTokenHeader(), handlers.AccessClusterStatus(apiClient))
	app.GET("/cluster/:cluster_id/kubeconfig", handlers.AddTokenHeader(), handlers.DownloadKubeconfigHandler(apiClient))

	// logout
	app.GET("/logout", handlers.Logout())
}
