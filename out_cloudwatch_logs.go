package main

import "github.com/fluent/fluent-bit-go/output"
import "github.com/json-iterator/go"
import "github.com/aws/aws-sdk-go/aws"
import "github.com/aws/aws-sdk-go/aws/awserr"
import "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
import "github.com/aws/aws-sdk-go/aws/session"

import (
	"C"
	"fmt"
	"os"
	"time"
	"unsafe"
)

var plugin GoOutputPlugin = &fluentPlugin{}
var cloudwatchLogs *cloudwatchlogs.CloudWatchLogs

type cloudWatchLogsConf struct {
	logGroupName     string
	logStreamName    string
	autoCreateStream bool
}

type updateToken struct {
	logGroup  string
	logStream string
}

var configCtx *cloudWatchLogsConf
var sequenceTokensCtx map[updateToken]string

type GoOutputPlugin interface {
	PluginConfigKey(ctx unsafe.Pointer, key string) string
	Unregister(ctx unsafe.Pointer)
	GetRecord(dec *output.FLBDecoder) (ret int, ts interface{}, rec map[interface{}]interface{})
	NewDecoder(data unsafe.Pointer, length int) *output.FLBDecoder
	Put(timestamp time.Time, line, sequenceToken string) (*cloudwatchlogs.PutLogEventsOutput, error)
	CheckLogGroupsExistence(logGroupName string) bool
	CheckLogStreamsExistence(logGroupName, logStreamName string) (bool, string)
	CreateLogGroup(logGroupName string) error
	CreateLogStream(logGroupName, logStreamName string) error
	Exit(code int)
}

type fluentPlugin struct{}

func (p *fluentPlugin) PluginConfigKey(ctx unsafe.Pointer, key string) string {
	return output.FLBPluginConfigKey(ctx, key)
}

func (p *fluentPlugin) Unregister(ctx unsafe.Pointer) {
	output.FLBPluginUnregister(ctx)
}

func (p *fluentPlugin) GetRecord(dec *output.FLBDecoder) (int, interface{}, map[interface{}]interface{}) {
	return output.GetRecord(dec)
}

func (p *fluentPlugin) NewDecoder(data unsafe.Pointer, length int) *output.FLBDecoder {
	return output.NewDecoder(data, int(length))
}

func (p *fluentPlugin) Exit(code int) {
	os.Exit(code)
}

func (p *fluentPlugin) Put(timestamp time.Time, line string, sequenceToken string) (*cloudwatchlogs.PutLogEventsOutput, error) {
	t := aws.TimeUnixMilli(timestamp)
	params := &cloudwatchlogs.PutLogEventsInput{
		LogEvents: []*cloudwatchlogs.InputLogEvent{ // Mandatory
			&cloudwatchlogs.InputLogEvent{ // Mandatory
				Message:   aws.String(line), // Mandatory
				Timestamp: aws.Int64(t),     // Mandatory
			},
			// More values
		},
		LogGroupName:  aws.String(configCtx.logGroupName),  // Mandatory
		LogStreamName: aws.String(configCtx.logStreamName), // Mandatory
	}
	if sequenceToken != "" {
		params.SequenceToken = aws.String(sequenceToken)
	}
	resp, err := cloudwatchLogs.PutLogEvents(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Get error details
			fmt.Println("Error:", awsErr.Code(), awsErr.Message())

			// Prints out full error message, including original error if there was one.
			fmt.Println("Error:", awsErr.Error())
			return nil, err
		} else if err != nil {
			// A non-service error occurred.
			// A service error occurred.
			fmt.Printf("Fatal: %v\n", err)
			return nil, err
		}
	}

	return resp, nil
}

func (p *fluentPlugin) CheckLogGroupsExistence(logGroupName string) bool {
	params := &cloudwatchlogs.DescribeLogGroupsInput{
		LogGroupNamePrefix: aws.String(logGroupName), // Required
	}
	resp, err := cloudwatchLogs.DescribeLogGroups(params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false
	}

	logGroups := resp.LogGroups

	for _, logGroup := range logGroups {
		if logGroupName == *logGroup.LogGroupName {
			return true
		}
	}

	return false
}

func (p *fluentPlugin) CheckLogStreamsExistence(logGroupName, logStreamName string) (bool, string) {
	params := &cloudwatchlogs.DescribeLogStreamsInput{
		LogGroupName:        aws.String(logGroupName),
		LogStreamNamePrefix: aws.String(logStreamName), // Required
	}
	resp, err := cloudwatchLogs.DescribeLogStreams(params)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return false, ""
	}

	logStreams := resp.LogStreams

	for _, logStream := range logStreams {
		if logStreamName == *logStream.LogStreamName {
			nextToken := *logStream.UploadSequenceToken
			if nextToken != "" {
				return true, nextToken
			} else {
				return false, ""
			}
		}
	}

	return false, ""
}

func (p *fluentPlugin) CreateLogGroup(logGroupName string) error {
	params := &cloudwatchlogs.CreateLogGroupInput{
		LogGroupName: aws.String(logGroupName), // Required
	}
	_, err := cloudwatchLogs.CreateLogGroup(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Get error details
			fmt.Println("Error:", awsErr.Code(), awsErr.Message())

			// Prints out full error message, including original error if there was one.
			fmt.Println("Error:", awsErr.Error())
			return err
		} else if err != nil {
			// A non-service error occurred.
			// A service error occurred.
			fmt.Printf("Fatal: %v\n", err)
			return err
		}
	}

	return nil
}

func (p *fluentPlugin) CreateLogStream(logGroupName, logStreamName string) error {
	params := &cloudwatchlogs.CreateLogStreamInput{
		LogGroupName:  aws.String(logGroupName),  // Required
		LogStreamName: aws.String(logStreamName), // Required
	}
	_, err := cloudwatchLogs.CreateLogStream(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			// Get error details
			fmt.Println("Error:", awsErr.Code(), awsErr.Message())

			// Prints out full error message, including original error if there was one.
			fmt.Println("Error:", awsErr.Error())
			return err
		} else if err != nil {
			// A non-service error occurred.
			// A service error occurred.
			fmt.Printf("Fatal: %v\n", err)
			return err
		}
	}

	return nil
}

//export FLBPluginRegister
func FLBPluginRegister(ctx unsafe.Pointer) int {
	return output.FLBPluginRegister(ctx, "cloudwatch_logs", "ClooudwatchLogs Output plugin written in GO!")
}

//export FLBPluginInit
// (fluentbit will call this)
// ctx (context) pointer to fluentbit context (state/ c code)
func FLBPluginInit(ctx unsafe.Pointer) int {
	// Example to retrieve an optional configuration parameter
	credential := plugin.PluginConfigKey(ctx, "Credential")
	accessKeyID := plugin.PluginConfigKey(ctx, "AccessKeyID")
	secretAccessKey := plugin.PluginConfigKey(ctx, "SecretAccessKey")
	logStreamName := plugin.PluginConfigKey(ctx, "LogGroupName")
	logGroupName := plugin.PluginConfigKey(ctx, "LogStreamName")
	region := plugin.PluginConfigKey(ctx, "Region")
	autoCreateStream := plugin.PluginConfigKey(ctx, "AutoCreateStream")

	config, err := getCloudWatchLogsConfig(accessKeyID, secretAccessKey, credential, logGroupName, logStreamName, region, autoCreateStream)
	if err != nil {
		plugin.Unregister(ctx)
		plugin.Exit(1)
		return output.FLB_ERROR
	}
	fmt.Printf("[flb-go] plugin credential parameter = '%s'\n", credential)
	fmt.Printf("[flb-go] plugin accessKeyID parameter = '%s'\n", secretConfig(accessKeyID))
	fmt.Printf("[flb-go] plugin secretAccessKey parameter = '%s'\n", secretConfig(secretAccessKey))
	fmt.Printf("[flb-go] plugin logGroupName parameter = '%s'\n", logGroupName)
	fmt.Printf("[flb-go] plugin logStreamName parameter = '%s'\n", logStreamName)
	fmt.Printf("[flb-go] plugin region parameter = '%s'\n", region)
	fmt.Printf("[flb-go] plugin autoCreateStream parameter = '%s'\n", autoCreateStream)

	sess := session.New(&aws.Config{
		Credentials: config.credentials,
		Region:      config.region,
	})
	cloudwatchLogs = cloudwatchlogs.New(sess)

	configCtx = &cloudWatchLogsConf{
		logGroupName:     *config.logGroupName,
		logStreamName:    *config.logStreamName,
		autoCreateStream: config.autoCreateStream,
	}

	if configCtx.autoCreateStream {
		if !plugin.CheckLogGroupsExistence(configCtx.logGroupName) {
			err := plugin.CreateLogGroup(configCtx.logGroupName)
			if err != nil {
				fmt.Printf("Failed to create logGroup. error: %v\n", err)
			}
		}
	}

	if configCtx.autoCreateStream {
		if doesExist, nextToken := plugin.CheckLogStreamsExistence(configCtx.logGroupName, configCtx.logStreamName); doesExist {
			sequenceTokensCtx = make(map[updateToken]string)
			sequenceTokensCtx[updateToken{configCtx.logGroupName, configCtx.logStreamName}] = nextToken
		} else {
			err := plugin.CreateLogStream(configCtx.logGroupName, configCtx.logStreamName)
			if err != nil {
				fmt.Printf("Failed to create logStream. error: %v\n", err)
			}
		}
	}

	return output.FLB_OK
}

//export FLBPluginFlush
func FLBPluginFlush(data unsafe.Pointer, length C.int, tag *C.char) int {
	var ret int
	var ts interface{}
	var record map[interface{}]interface{}

	dec := plugin.NewDecoder(data, int(length))

	for {
		ret, ts, record = plugin.GetRecord(dec)
		if ret != 0 {
			break
		}

		// Get timestamp
		var timestamp time.Time
		switch t := ts.(type) {
		case output.FLBTime:
			timestamp = ts.(output.FLBTime).Time
		case uint64:
			timestamp = time.Unix(int64(t), 0)
		default:
			fmt.Print("timestamp isn't known format. Use current time.\n")
			timestamp = time.Now()
		}

		line, err := createJSON(record)
		if err != nil {
			fmt.Printf("error creating message for S3: %v\n", err)
			continue
		}

		resp, err := plugin.Put(timestamp, line, sequenceTokensCtx[updateToken{configCtx.logGroupName, configCtx.logStreamName}])
		if err != nil {
			fmt.Printf("error sending message for S3: %v\n", err)
			return output.FLB_RETRY
		}
		sequenceTokensCtx[updateToken{configCtx.logGroupName, configCtx.logStreamName}] = nextSequenceToken(resp)
	}

	// Return options:
	//
	// output.FLB_OK    = data have been processed.
	// output.FLB_ERROR = unrecoverable error, do not try this again.
	// output.FLB_RETRY = retry to flush later.
	return output.FLB_OK
}

func secretConfig(parameter string) string {
	if parameter != "" {
		return "xxxxxx"
	} else {
		return ""
	}
}

func nextSequenceToken(response *cloudwatchlogs.PutLogEventsOutput) string {
	if response != nil {
		return *response.NextSequenceToken
	} else {
		return ""
	}
}

func createJSON(record map[interface{}]interface{}) (string, error) {
	m := make(map[string]interface{})

	for k, v := range record {
		switch t := v.(type) {
		case []byte:
			// prevent encoding to base64
			m[k.(string)] = string(t)
		default:
			m[k.(string)] = v
		}
	}

	js, err := jsoniter.Marshal(m)
	if err != nil {
		return "{}", err
	}

	return string(js), nil
}

//export FLBPluginExit
func FLBPluginExit() int {
	return output.FLB_OK
}

func main() {
}
