package main

import (
	"html/template"
	"io"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"bitbucket.com/turntwo/quicksight-embeds/config"
	m "bitbucket.com/turntwo/quicksight-embeds/middleware"
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

	cfg := config.New()
	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionKey))
	e.Use(session.Middleware(sessionStore))
	e.Use(middleware.Logger())

	handler := web.NewHandler(cfg)
	ah := web.AuthHandler{handler}
	dh := web.DashboardHandler{handler}

	// Authentication Handlers
	e.GET("/", ah.Index)
	e.POST("/authenticate", ah.SubmitLogin)

	// Dashboard handlers
	e.GET("/dashboard", dh.Index, m.CognitoAuthentication(sessionStore, cfg))

	e.Logger.Fatal(e.Start(":1323"))
}
