package main

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type LoginRequest struct {
	Email    string `json:"email" form:"email" query:"email"`
	Password string `json:"password" form:"password" query:"password"`
}

func RenderLogin(c echo.Context) error {
	return c.Render(http.StatusOK, "login", map[string]interface{}{})
}

func SubmitLogin(c echo.Context) error {
	request := new(LoginRequest)
	if err := c.Bind(request); err != nil {
		return err
	}

	// Dry this up
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	cfg.Region = endpoints.ApSoutheast1RegionID

	cognito := cognitoidentityprovider.New(cfg)
	req := cognito.InitiateAuthRequest(&cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: cognitoidentityprovider.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String("70smiki89gas0prcmog2vio1v5"),
		AuthParameters: map[string]string{
			"USERNAME": request.Email,
			"PASSWORD": request.Password,
		},
	})
	resp, err := req.Send(context.TODO())
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err)
	}
	fmt.Println(resp)

	cfg.Region = endpoints.UsEast1RegionID
	qs := quicksight.New(cfg)
	qsRequest := qs.DescribeUserRequest(&quicksight.DescribeUserInput{
		AwsAccountId: aws.String("259921281957"),
		Namespace:    aws.String("default"),
		UserName:     aws.String("QuickSightReaderRole/" + request.Email),
	})
	_, err = qsRequest.Send(context.TODO())

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case quicksight.ErrCodeResourceNotFoundException:
				registerReq := qs.RegisterUserRequest(&quicksight.RegisterUserInput{
					AwsAccountId: aws.String("259921281957"),
					IamArn:       aws.String("arn:aws:iam::259921281957:role/QuickSightReaderRole"),
					IdentityType: quicksight.IdentityTypeIam,
					Namespace:    aws.String("default"),
					Email:        aws.String(request.Email),
					SessionName:  aws.String(request.Email),
					UserRole:     quicksight.UserRoleReader,
				})
				qsRegisterResp, err := registerReq.Send(context.TODO())
				if err != nil {
					fmt.Println(err)
				} else {
					groupReq := qs.CreateGroupMembershipRequest(&quicksight.CreateGroupMembershipInput{
						AwsAccountId: aws.String("259921281957"),
						GroupName:    aws.String("embed-readers"),
						Namespace:    aws.String("default"),
						MemberName:   qsRegisterResp.User.UserName,
					})
					qsGroupResp, err := groupReq.Send(context.TODO())
					fmt.Println(qsGroupResp, err)
				}
			default:
				fmt.Println(err)
			}
		} else {
			fmt.Println(err)
		}
	}

	cfg.Region = endpoints.ApSoutheast1RegionID
	stsClient := sts.New(cfg)
	stsReq := stsClient.AssumeRoleRequest(&sts.AssumeRoleInput{
		RoleArn:         aws.String("arn:aws:iam::259921281957:role/QuickSightReaderRole"),
		RoleSessionName: aws.String(request.Email),
	})
	stsResp, err := stsReq.Send(context.TODO())
	if err != nil {
		fmt.Println(err)
	}

	cfgCopy := cfg.Copy()
	cfgCopy.Credentials = CredentialsProvider{Credentials: stsResp.Credentials}
	cfgCopy.Region = endpoints.ApSoutheast1RegionID

	qs = quicksight.New(cfgCopy)
	qsEmbedReq := qs.GetDashboardEmbedUrlRequest(&quicksight.GetDashboardEmbedUrlInput{
		AwsAccountId: aws.String("259921281957"),
		DashboardId:  aws.String("cbda847a-4098-4cd6-ba25-23037b9b4586"),
		IdentityType: quicksight.IdentityTypeIam,
	})
	qsEmbedResp, err := qsEmbedReq.Send(context.TODO())
	if err != nil {
		fmt.Println(err)
	}

	return c.JSON(http.StatusOK, qsEmbedResp)
}

type CredentialsProvider struct {
	*sts.Credentials
}

func (s CredentialsProvider) Retrieve() (aws.Credentials, error) {

	if s.Credentials == nil {
		return aws.Credentials{}, errors.New("sts credentials are nil")
	}

	return aws.Credentials{
		AccessKeyID:     aws.StringValue(s.AccessKeyId),
		SecretAccessKey: aws.StringValue(s.SecretAccessKey),
		SessionToken:    aws.StringValue(s.SessionToken),
		Expires:         aws.TimeValue(s.Expiration),
	}, nil
}

func main() {
	e := echo.New()
	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}
	e.Renderer = t
	e.Static("/assets", "public/assets")
	e.Use(middleware.Logger())

	e.GET("/", RenderLogin)
	e.POST("/authenticate", SubmitLogin)
	e.Logger.Fatal(e.Start(":1323"))
}
