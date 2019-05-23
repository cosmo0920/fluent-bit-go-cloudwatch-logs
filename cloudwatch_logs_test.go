package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	// "time"
)

func TestGetS3ConfigStaticCredentials(t *testing.T) {
	conf, err := getCloudWatchLogsConfig("exampleaccessID", "examplesecretkey", "", "examplelogGroup", "exampleLogstream", "exampleregion", "", "")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	assert.Equal(t, "examplelogGroup", *conf.logGroupName, "Specify logGroup name")
	assert.Equal(t, "exampleLogstream", *conf.logStreamName, "Specify logStream name")
	assert.NotNil(t, conf.credentials, "credentials not to be nil")
	assert.Equal(t, "exampleregion", *conf.region, "Specify region name")
	assert.Equal(t, true, conf.autoCreateStream, "Specify autocreatestream flag")
	assert.Equal(t, "", *conf.stateFile, "Specify stateFile path correctly")
}

func TestGetS3ConfigSharedCredentials(t *testing.T) {
	cloudwatchLogsCreds = &testCloudwatchLogsCredential{}
	conf, err := getCloudWatchLogsConfig("", "", "examplecredentials", "examplelogGroup", "exampleLogstream", "exampleregion", "", "")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	assert.Equal(t, "examplelogGroup", *conf.logGroupName, "Specify logGroup name")
	assert.Equal(t, "exampleLogstream", *conf.logStreamName, "Specify logStream name")
	assert.NotNil(t, conf.credentials, "credentials not to be nil")
	assert.Equal(t, "exampleregion", *conf.region, "Specify region name")
	assert.Equal(t, true, conf.autoCreateStream, "Specify autocreatestream flag")
	assert.Equal(t, "", *conf.stateFile, "Specify stateFile path correctly")
}

func TestStateFileFor(t *testing.T) {
	conf, err := getCloudWatchLogsConfig("exampleaccessID", "examplesecretkey", "", "examplelogGroup", "example/Log/stream", "exampleregion", "", "examplestatefile")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}

	result := stateFileFor(*conf.stateFile, *conf.logStreamName)
	assert.Equal(t, "examplestatefile_example-Log-stream", result, "Created stateFile path wrongly")
}

func TestStateFileReadWrite(t *testing.T) {
	sequenceTokenCtx = ""
	conf, err := getCloudWatchLogsConfig("exampleaccessID", "examplesecretkey", "", "examplelogGroup", "example/Log/stream", "exampleregion", "", "examplestatefile")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	stateFile := stateFileFor(*conf.stateFile, *conf.logStreamName)

	nextSequenceToken := "1234567890"
	storeStateToken("examplestatefile", "example/Log/stream", nextSequenceToken)
	result := readStateToken("examplestatefile", "example/Log/stream")

	defer os.Remove(stateFile)

	fmt.Printf("nextStateToken: %s\n", nextSequenceToken)
	assert.Equal(t, nextSequenceToken, result, "Created stateFile path wrongly")

}
