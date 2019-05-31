package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/labstack/echo/v3"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func Login(c echo.Context) error {
	return c.Render(http.StatusOK, "login", map[string]interface{}{})
}

func main() {
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	cfg.Region = endpoints.ApSoutheast1RegionID

	client := quicksight.New(cfg)
	req := client.GetDashboardEmbedUrlRequest(&quicksight.GetDashboardEmbedUrlInput{
		AwsAccountId: aws.String("259921281957"),
		DashboardId:  aws.String("6e38bf60-a417-4817-a154-8a5a5040278e"),
		IdentityType: quicksight.IdentityTypeIam,
	})
	resp, err := req.Send(context.TODO())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp)

	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}
	e.Renderer = t
	e.Static("/assets", "public/assets")

	e.GET("/", Login)
	e.Logger.Fatal(e.Start(":1323"))
}
