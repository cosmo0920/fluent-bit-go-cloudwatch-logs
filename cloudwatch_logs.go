package main

import "github.com/aws/aws-sdk-go/aws"
import "github.com/aws/aws-sdk-go/aws/credentials"

import (
	"fmt"
	"strconv"
)

type cloudwatchLogsConfig struct {
	credentials      *credentials.Credentials
	logGroupName     *string
	logStreamName    *string
	region           *string
	autoCreateStream bool
	stateFile        *string
}

type CloudWatchLogsCredential interface {
	GetCredentials(accessID, secretkey, credentials string) (*credentials.Credentials, error)
}

type cloudwatchLogsPluginConfig struct{}

var cloudwatchLogsCreds CloudWatchLogsCredential = &cloudwatchLogsPluginConfig{}

func (c *cloudwatchLogsPluginConfig) GetCredentials(accessKeyID, secretKey, credential string) (*credentials.Credentials, error) {
	var creds *credentials.Credentials
	if credential != "" {
		creds = credentials.NewSharedCredentials(credential, "default")
		if _, err := creds.Get(); err != nil {
			fmt.Println("[SharedCredentials] ERROR:", err)
		} else {
			return creds, nil
		}
	} else if !(accessKeyID == "" && secretKey == "") {
		creds = credentials.NewStaticCredentials(accessKeyID, secretKey, "")
		if _, err := creds.Get(); err != nil {
			fmt.Println("[StaticCredentials] ERROR:", err)
		} else {
			return creds, nil
		}
	} else {
		creds = credentials.NewEnvCredentials()
		if _, err := creds.Get(); err != nil {
			fmt.Println("[EnvCredentials] ERROR:", err)
		} else {
			return creds, nil
		}
	}

	return nil, fmt.Errorf("Failed to create credentials")
}

func getCloudWatchLogsConfig(accessID, secretKey, credential, logGroupName, logStreamName, region, autoCreateStream, stateFile string) (*cloudwatchLogsConfig, error) {
	conf := &cloudwatchLogsConfig{}
	creds, err := cloudwatchLogsCreds.GetCredentials(accessID, secretKey, credential)
	if err != nil {
		return nil, fmt.Errorf("Failed to create credentials")
	}
	conf.credentials = creds

	if logGroupName == "" {
		return nil, fmt.Errorf("Cannot specify empty string to bucket name")
	}
	conf.logGroupName = aws.String(logGroupName)

	if logStreamName == "" {
		return nil, fmt.Errorf("Cannot specify empty string to logStreamName")
	}
	conf.logStreamName = aws.String(logStreamName)

	if region == "" {
		return nil, fmt.Errorf("Cannot specify empty string to region")
	}
	conf.region = aws.String(region)

	if autoCreateStream == "" {
		conf.autoCreateStream = true
	}
	ok, err := strconv.ParseBool(autoCreateStream)
	if err != nil {
		conf.autoCreateStream = true
	} else {
		conf.autoCreateStream = ok
	}

	// optional
	if stateFile == "" {
		conf.stateFile = aws.String("")
	} else {
		conf.stateFile = aws.String(stateFile)
	}

	return conf, nil
}
