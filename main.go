package main

import (
	"html/template"
	"io"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"bitbucket.com/turntwo/quicksight-embeds/config"
	"bitbucket.com/turntwo/quicksight-embeds/web"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	e.Renderer = &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}
	e.Static("/assets", "public/assets")
	e.Use(middleware.Logger())

	handler := web.NewHandler(config.New())
	ah := web.AuthHandler{handler}

	// Authentication Handlers
	e.GET("/", ah.RenderLogin)
	e.POST("/authenticate", ah.SubmitLogin)

	e.Logger.Fatal(e.Start(":1323"))
}
