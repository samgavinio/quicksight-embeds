package config

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/endpoints"
	"github.com/aws/aws-sdk-go-v2/aws/external"
	"github.com/spf13/viper"
)

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

func New() *Config {
	// Viper support for both environment variables and configuration file
	viper := viper.GetViper()
	viper.AutomaticEnv()
	viper.SetConfigName("config")
	viper.AddConfigPath("./config")
	viper.ReadInConfig()

	cfg, err := external.LoadDefaultAWSConfig()
	if err != nil {
		panic("unable to load SDK config, " + err.Error())
	}
	cfg.Region = endpoints.ApSoutheast1RegionID

	return &Config{
		AWS: AWS{
			AccountId: viper.GetString("AWS_ACCOUNT_ID"),
			Config:    cfg,
		},
		Cognito: Cognito{
			ClientId: viper.GetString("COGNITO_CLIENT_ID"),
		},
		Quicksight: Quicksight{
			RoleName:    viper.GetString("QUICKSIGHT_IAM_ROLE_NAME"),
			Group:       viper.GetString("QUICKSIGHT_GROUP_NAME"),
			DashboardId: viper.GetString("QUICKSIGHT_DASHBOARD_ID"),
			Namespace:   "default",
		},
	}
}
