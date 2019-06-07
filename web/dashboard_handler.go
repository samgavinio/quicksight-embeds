package web

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/service/quicksight"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type DashboardHandler struct {
	Handler
}

type CredentialsProvider struct {
	*sts.Credentials
}

func (h *DashboardHandler) Index(c echo.Context) (err error) {
	sess, _ := session.Get("session", c)
	email := sess.Values["user_email"].(string)

	// If EmbedUrl is in the session, use it instead
	var url string
	qsSessionKey := fmt.Sprintf("quicksight_embed_url_%s", email)
	if sess.Values[qsSessionKey] != nil {
		url = sess.Values[qsSessionKey].(string)
	} else {
		response, err := h.getDashboardUrl(email)
		if err != nil {
			fmt.Println(err)
			return c.JSON(http.StatusBadRequest, map[string]string{"Error": "Bad Request."})
		}
		url = *response.EmbedUrl

		sess.Values[qsSessionKey] = response.EmbedUrl
		sess.Save(c.Request(), c.Response())
	}

	return c.Render(http.StatusOK, "dashboard", map[string]interface{}{
		"QuicksightUrl": url,
	})
}

// Generates the Quicksight Embed URL. If the passed email does not yet exist in Quicksight a new user is provisioned.
func (h *DashboardHandler) getDashboardUrl(email string) (qsEmbedResp *quicksight.GetDashboardEmbedUrlResponse, err error) {
	useastCfg := h.Config.AWS.Config.Copy()
	useastCfg.Region = endpoints.UsEast1RegionID
	qs := quicksight.New(useastCfg)
	dRequest := qs.DescribeUserRequest(&quicksight.DescribeUserInput{
		AwsAccountId: aws.String(h.Config.AWS.AccountId),
		Namespace:    aws.String(h.Config.Quicksight.Namespace),
		UserName:     aws.String(fmt.Sprintf("%s/%s", h.Config.Quicksight.RoleName, email)),
	})

	if _, err = dRequest.Send(context.TODO()); err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			switch awsErr.Code() {
			case quicksight.ErrCodeResourceNotFoundException:
				registerReq := qs.RegisterUserRequest(&quicksight.RegisterUserInput{
					AwsAccountId: aws.String(h.Config.AWS.AccountId),
					IamArn:       aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", h.Config.AWS.AccountId, h.Config.Quicksight.RoleName)),
					IdentityType: quicksight.IdentityTypeIam,
					Namespace:    aws.String(h.Config.Quicksight.Namespace),
					Email:        aws.String(email),
					SessionName:  aws.String(email),
					UserRole:     quicksight.UserRoleReader,
				})
				if qsRegisterResp, err := registerReq.Send(context.TODO()); err != nil {
					groupReq := qs.CreateGroupMembershipRequest(&quicksight.CreateGroupMembershipInput{
						AwsAccountId: aws.String(h.Config.AWS.AccountId),
						GroupName:    aws.String(h.Config.Quicksight.Group),
						Namespace:    aws.String(h.Config.Quicksight.Namespace),
						MemberName:   qsRegisterResp.User.UserName,
					})
					_, err = groupReq.Send(context.TODO())
				}
			}
		}
	}

	stsClient := sts.New(h.Config.AWS.Config)
	stsReq := stsClient.AssumeRoleRequest(&sts.AssumeRoleInput{
		RoleArn:         aws.String(fmt.Sprintf("arn:aws:iam::%s:role/%s", h.Config.AWS.AccountId, h.Config.Quicksight.RoleName)),
		RoleSessionName: aws.String(email),
	})
	stsResp, err := stsReq.Send(context.TODO())

	cfgCopy := h.Config.AWS.Config.Copy()
	cfgCopy.Credentials = CredentialsProvider{Credentials: stsResp.Credentials}
	cfgCopy.Region = endpoints.ApSoutheast1RegionID

	qs = quicksight.New(cfgCopy)
	qsEmbedReq := qs.GetDashboardEmbedUrlRequest(&quicksight.GetDashboardEmbedUrlInput{
		AwsAccountId: aws.String(h.Config.AWS.AccountId),
		DashboardId:  aws.String(h.Config.Quicksight.DashboardId),
		IdentityType: quicksight.IdentityTypeIam,
	})
	qsEmbedResp, err = qsEmbedReq.Send(context.TODO())

	if err != nil {
		return nil, err
	}

	return qsEmbedResp, nil
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
