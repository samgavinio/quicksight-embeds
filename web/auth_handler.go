package web

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Handler
}

type loginRequest struct {
	Email    string `json:"email" form:"email" query:"email"`
	Password string `json:"password" form:"password" query:"password"`
}

func (h *AuthHandler) Index(c echo.Context) (err error) {
	h.clearSession(c)

	return c.Render(http.StatusOK, "login", map[string]interface{}{})
}

func (h *AuthHandler) Logout(c echo.Context) (err error) {
	h.clearSession(c)

	return c.Redirect(http.StatusMovedPermanently, "/")
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
		return c.Render(http.StatusUnauthorized, "login", map[string]interface{}{
			"Error": "Invalid username/password.",
		})
	}

	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600,
		HttpOnly: true,
	}
	sess.Values["cognito_access_token"] = resp.AuthenticationResult.AccessToken
	sess.Values["user_email"] = request.Email
	sess.Values[fmt.Sprintf("quicksight_embed_url_%s", request.Email)] = nil
	sess.Save(c.Request(), c.Response())

	return c.Redirect(http.StatusMovedPermanently, "/dashboard")
}

func (h *AuthHandler) clearSession(c echo.Context) {
	// Quick and dirty: simply clear the session
	sess, _ := session.Get("session", c)
	sess.Options = &sessions.Options{MaxAge: -1}
	sess.Save(c.Request(), c.Response())
}
