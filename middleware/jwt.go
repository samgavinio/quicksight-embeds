package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

// Author: yasu-murata <https://github.com/mura123yasu>
// Source: https://github.com/mura123yasu/go-cognito
// Validates the JWT Token obtained from Cognito.
type JWTValidator struct {
	Region            string
	CognitoUserPoolId string
}

// JWK is json data struct for JSON Web Key
type JWK struct {
	Keys []JWKKey
}

// JWKKey is json data struct for cognito jwk key
type JWKKey struct {
	Alg string
	E   string
	Kid string
	Kty string
	N   string
	Use string
}

func (j *JWTValidator) Validate(tokenStr string) (*jwt.Token, error) {
	// Download and store the JSON Web Key (JWK) for your user pool.
	// @todo: cache the results
	jwkURL := fmt.Sprintf("https://cognito-idp.%v.amazonaws.com/%v/.well-known/jwks.json", j.Region, j.CognitoUserPoolId)
	fmt.Println(jwkURL)
	jwk := getJWK(jwkURL)

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		if kid, ok := token.Header["kid"]; ok {
			if kidStr, ok := kid.(string); ok {
				key := jwk[kidStr]

				// Verify the signature of the decoded JWT token.
				rsaPublicKey := convertKey(key.E, key.N)
				return rsaPublicKey, nil
			}
		}

		return "", nil
	})

	if err != nil {
		return token, err
	}

	claims := token.Claims.(jwt.MapClaims)

	iss, ok := claims["iss"]
	if !ok {
		return token, fmt.Errorf("token does not contain issuer")
	}
	issStr := iss.(string)
	if strings.Contains(issStr, "cognito-idp") {
		err = validateAWSJwtClaims(claims, j.Region, j.CognitoUserPoolId)
		if err != nil {
			return token, err
		}
	}

	if token.Valid {
		return token, nil
	}
	return token, err
}

func getJSON(url string, target interface{}) error {
	var myClient = &http.Client{Timeout: 10 * time.Second}
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}

func getJWK(jwkURL string) map[string]JWKKey {

	jwk := &JWK{}

	getJSON(jwkURL, jwk)

	jwkMap := make(map[string]JWKKey, 0)
	for _, jwk := range jwk.Keys {
		jwkMap[jwk.Kid] = jwk
	}
	return jwkMap
}

func convertKey(rawE, rawN string) *rsa.PublicKey {
	decodedE, err := base64.RawURLEncoding.DecodeString(rawE)
	if err != nil {
		panic(err)
	}
	if len(decodedE) < 4 {
		ndata := make([]byte, 4)
		copy(ndata[4-len(decodedE):], decodedE)
		decodedE = ndata
	}
	pubKey := &rsa.PublicKey{
		N: &big.Int{},
		E: int(binary.BigEndian.Uint32(decodedE[:])),
	}
	decodedN, err := base64.RawURLEncoding.DecodeString(rawN)
	if err != nil {
		panic(err)
	}
	pubKey.N.SetBytes(decodedN)
	return pubKey
}

// validateAWSJwtClaims validates AWS Cognito User Pool JWT
func validateAWSJwtClaims(claims jwt.MapClaims, region, userPoolID string) error {
	var err error
	// Check the iss claim. It should match your user pool.
	issShoudBe := fmt.Sprintf("https://cognito-idp.%v.amazonaws.com/%v", region, userPoolID)
	err = validateClaimItem("iss", []string{issShoudBe}, claims)
	if err != nil {
		return err
	}

	// Check the token_use claim.
	validateTokenUse := func() error {
		if tokenUse, ok := claims["token_use"]; ok {
			if tokenUseStr, ok := tokenUse.(string); ok {
				if tokenUseStr == "id" || tokenUseStr == "access" {
					return nil
				}
			}
		}
		return errors.New("token_use should be id or access")
	}

	err = validateTokenUse()
	if err != nil {
		return err
	}

	// Check the exp claim and make sure the token is not expired.
	err = validateExpired(claims)
	if err != nil {
		return err
	}

	return nil
}

func validateClaimItem(key string, keyShouldBe []string, claims jwt.MapClaims) error {
	if val, ok := claims[key]; ok {
		if valStr, ok := val.(string); ok {
			for _, shouldbe := range keyShouldBe {
				if valStr == shouldbe {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("%v does not match any of valid values: %v", key, keyShouldBe)
}

func validateExpired(claims jwt.MapClaims) error {
	if tokenExp, ok := claims["exp"]; ok {
		if exp, ok := tokenExp.(float64); ok {
			now := time.Now().Unix()
			fmt.Printf("current unixtime : %v\n", now)
			fmt.Printf("expire unixtime  : %v\n", int64(exp))
			if int64(exp) > now {
				return nil
			}
		}
		return errors.New("cannot parse token exp")
	}
	return errors.New("token is expired")
}
