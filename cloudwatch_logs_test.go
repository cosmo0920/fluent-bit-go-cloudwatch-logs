package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
	// "time"
)

func TestGetS3ConfigStaticCredentials(t *testing.T) {
	conf, err := getCloudWatchLogsConfig("exampleaccessID", "examplesecretkey", "", "examplelogGroup", "exampleLogstream", "exampleregion", "")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	assert.Equal(t, "examplelogGroup", *conf.logGroupName, "Specify logGroup name")
	assert.Equal(t, "exampleLogstream", *conf.logStreamName, "Specify logStream name")
	assert.NotNil(t, conf.credentials, "credentials not to be nil")
	assert.Equal(t, "exampleregion", *conf.region, "Specify region name")
	assert.Equal(t, true, conf.autoCreateStream, "Specify autocreatestream flag")
}

func TestGetS3ConfigSharedCredentials(t *testing.T) {
	cloudwatchLogsCreds = &testCloudwatchLogsCredential{}
	conf, err := getCloudWatchLogsConfig("", "", "examplecredentials", "examplelogGroup", "exampleLogstream", "exampleregion", "")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	assert.Equal(t, "examplelogGroup", *conf.logGroupName, "Specify logGroup name")
	assert.Equal(t, "exampleLogstream", *conf.logStreamName, "Specify logStream name")
	assert.NotNil(t, conf.credentials, "credentials not to be nil")
	assert.Equal(t, "exampleregion", *conf.region, "Specify region name")
	assert.Equal(t, true, conf.autoCreateStream, "Specify autocreatestream flag")
}
