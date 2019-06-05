package config

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/spf13/viper"
)

type Config struct {
	AWS
	Cognito
	Quicksight
	SessionKey string
}

type AWS struct {
	aws.Config
	AccountId string
	Region    string
}

type Cognito struct {
	ClientId   string
	UserPoolId string
}

type Quicksight struct {
	RoleName    string
	Group       string
	Namespace   string
	DashboardId string
}

func New() *Config {
	// Viper support for both environment variables and configuration file
	viper := viper.GetViper()
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.ReadInConfig()

	region := viper.GetString("AWS_REGION")
	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	cfg.Region = region
	cfg.LogLevel = aws.LogDebug

	return &Config{
		SessionKey: viper.GetString("SESSION_KEY"),
		AWS: AWS{
			AccountId: viper.GetString("AWS_ACCOUNT_ID"),
			Region:    region,
			Config:    cfg,
		},
		Cognito: Cognito{
			ClientId:   viper.GetString("COGNITO_CLIENT_ID"),
			UserPoolId: viper.GetString("COGNITO_USER_POOL_ID"),
		},
		Quicksight: Quicksight{
			RoleName:    viper.GetString("QUICKSIGHT_IAM_ROLE_NAME"),
			Group:       viper.GetString("QUICKSIGHT_GROUP_NAME"),
			DashboardId: viper.GetString("QUICKSIGHT_DASHBOARD_ID"),
			Namespace:   "default",
		},
	}
}
