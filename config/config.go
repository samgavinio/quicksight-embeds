package config

import "github.com/aws/aws-sdk-go-v2/aws"

type Config struct {
	AWS
	Cognito
	Quicksight
}

type AWS struct {
	aws.Config
	AccountId string
}

type Cognito struct {
	ClientId string
}

type Quicksight struct {
	RoleName    string
	Group       string
	Namespace   string
	DashboardId string
}
