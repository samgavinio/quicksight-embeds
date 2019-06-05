package main

import (
	"html/template"
	"io"

	"bitbucket.com/turntwo/quicksight-embeds/config"
	"bitbucket.com/turntwo/quicksight-embeds/web"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	cfg.Region = endpoints.ApSoutheast1RegionID

	handler := web.NewHandler(config.Config{
		AWS: config.AWS{
			AccountId: "259921281957",
			Config:    cfg,
		},
		Cognito: config.Cognito{
			ClientId: "70smiki89gas0prcmog2vio1v5",
		},
		Quicksight: config.Quicksight{
			RoleName:    "QuickSightReaderRole",
			Group:       "embed-readers",
			DashboardId: "cbda847a-4098-4cd6-ba25-23037b9b4586",
		},
	})
	ah := web.AuthHandler{handler}

	// Authentication Handlers
	e.GET("/", ah.RenderLogin)
	e.POST("/authenticate", ah.SubmitLogin)

	e.Logger.Fatal(e.Start(":1323"))
}
