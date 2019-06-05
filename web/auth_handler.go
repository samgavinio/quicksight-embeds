package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Handler
}

type loginRequest struct {
	Email    string `json:"email" form:"email" query:"email"`
	Password string `json:"password" form:"password" query:"password"`
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

func (h *AuthHandler) RenderLogin(c echo.Context) (err error) {
	return c.Render(http.StatusOK, "login", map[string]interface{}{})
}

func (h *AuthHandler) SubmitLogin(c echo.Context) (err error) {
	request := new(loginRequest)
	if err := c.Bind(request); err != nil {
		return err
	}

	cognito := cognitoidentityprovider.New(h.Config.AWS.Config)
	req := cognito.InitiateAuthRequest(&cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: cognitoidentityprovider.AuthFlowTypeUserPasswordAuth,
		ClientId: aws.String(h.Config.Cognito.ClientId),
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

	useastCfg := h.Config.AWS.Config.Copy()
	useastCfg.Region = endpoints.UsEast1RegionID
	qs := quicksight.New(useastCfg)
	qsRequest := qs.DescribeUserRequest(&quicksight.DescribeUserInput{
		AwsAccountId: aws.String(h.Config.AWS.AccountId),
		Namespace:    aws.String(h.Config.Quicksight.Namespace),
		UserName:     aws.String(fmt.Sprintf("%s/%s", h.Config.Quicksight.RoleName, request.Email)),
	})
	_, err = qsRequest.Send(context.TODO())

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case quicksight.ErrCodeResourceNotFoundException:
				registerReq := qs.RegisterUserRequest(&quicksight.RegisterUserInput{
					AwsAccountId: aws.String(h.Config.AWS.AccountId),
					IamArn:       aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", h.Config.AWS.AccountId, h.Config.Quicksight.RoleName)),
					IdentityType: quicksight.IdentityTypeIam,
					Namespace:    aws.String(h.Config.Quicksight.Namespace),
					Email:        aws.String(request.Email),
					SessionName:  aws.String(request.Email),
					UserRole:     quicksight.UserRoleReader,
				})
				qsRegisterResp, err := registerReq.Send(context.TODO())
				if err != nil {
					fmt.Println(err)
				} else {
					groupReq := qs.CreateGroupMembershipRequest(&quicksight.CreateGroupMembershipInput{
						AwsAccountId: aws.String(h.Config.AWS.AccountId),
						GroupName:    aws.String(h.Config.Quicksight.Group),
						Namespace:    aws.String(h.Config.Quicksight.Namespace),
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

	stsClient := sts.New(h.Config.AWS.Config)
	stsReq := stsClient.AssumeRoleRequest(&sts.AssumeRoleInput{
		RoleArn:         aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", h.Config.AWS.AccountId, h.Config.Quicksight.RoleName)),
		RoleSessionName: aws.String(request.Email),
	})
	stsResp, err := stsReq.Send(context.TODO())
	if err != nil {
		fmt.Println(err)
	}

	cfgCopy := h.Config.AWS.Config.Copy()
	cfgCopy.Credentials = CredentialsProvider{Credentials: stsResp.Credentials}
	cfgCopy.Region = endpoints.ApSoutheast1RegionID

	qs = quicksight.New(cfgCopy)
	qsEmbedReq := qs.GetDashboardEmbedUrlRequest(&quicksight.GetDashboardEmbedUrlInput{
		AwsAccountId: aws.String(h.Config.AWS.AccountId),
		DashboardId:  aws.String(h.Config.Quicksight.DashboardId),
		IdentityType: quicksight.IdentityTypeIam,
	})
	qsEmbedResp, err := qsEmbedReq.Send(context.TODO())
	if err != nil {
		fmt.Println(err)
	}

	return c.JSON(http.StatusOK, qsEmbedResp)
}
