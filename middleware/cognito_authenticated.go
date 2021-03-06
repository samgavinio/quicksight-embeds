package middleware

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"

	"bitbucket.com/turntwo/quicksight-embeds/config"
)

func onAuthFailure(c echo.Context) error {
	return c.Redirect(http.StatusMovedPermanently, "/")
}

func CognitoAuthentication(store *sessions.CookieStore, cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			request := c.Request()

			if sess, err := store.Get(request, "session"); err != nil {
				fmt.Println(err)
				return echo.ErrInternalServerError
			} else if sess.Values["cognito_access_token"] == nil || sess.Values["user_email"] == nil {
				fmt.Println("Current session is not authenticated.")
				return onAuthFailure(c)
			} else {
				v := JWTValidator{
					Region:            cfg.AWS.Region,
					CognitoUserPoolId: cfg.Cognito.UserPoolId,
				}
				token, err := v.Validate(sess.Values["cognito_access_token"].(string))
				if err != nil || !token.Valid {
					fmt.Printf("Token is not valid: %v\n", err)
					return onAuthFailure(c)
				}
			}
			return next(c)
		}
	}
}
