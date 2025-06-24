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

	//
	app.GET("/", handlers.RedirectToLogin())

	// login
	app.GET("/login", handlers.LoginPageHandlers())
	app.POST("/login", handlers.LoginHandlers(apiClient))

	// dashboard
	app.GET("/dashboard", handlers.AddTokenHeader(), handlers.DashboardPageHandlers(apiClient))

	// logout
	app.GET("/logout", handlers.Logout())
}
