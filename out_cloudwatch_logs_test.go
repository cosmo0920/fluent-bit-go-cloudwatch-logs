package main

import (
	"encoding/json"
	"testing"
	"time"
	"unsafe"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/fluent/fluent-bit-go/output"
	"github.com/stretchr/testify/assert"
)

func TestCreateJSON(t *testing.T) {
	record := make(map[interface{}]interface{})
	record["key"] = "value"
	record["number"] = 8

	line, err := createJSON(record)
	if err != nil {
		assert.Fail(t, "createJSON fails:%v", err)
	}
	assert.NotNil(t, line, "json string not to be nil")
	result := make(map[string]interface{})
	jsonBytes := ([]byte)(line)
	err = json.Unmarshal(jsonBytes, &result)
	if err != nil {
		assert.Fail(t, "unmarshal of json fails:%v", err)
	}

	assert.Equal(t, result["key"], "value")
	assert.Equal(t, result["number"], float64(8))
}

type testrecord struct {
	rc   int
	ts   interface{}
	data map[interface{}]interface{}
}

type events struct {
	data []byte
}
type testFluentPlugin struct {
	credential       string
	accessKeyID      string
	secretAccessKey  string
	logGroupName     string
	logStreamName    string
	region           string
	autoCreateStream string
	stateFile        string
	records          []testrecord
	position         int
	events           []*events
}

func (p *testFluentPlugin) PluginConfigKey(ctx unsafe.Pointer, key string) string {
	switch key {
	case "Credential":
		return p.credential
	case "AccessKeyID":
		return p.accessKeyID
	case "SecretAccessKey":
		return p.secretAccessKey
	case "LogGroupName":
		return p.logGroupName
	case "LogStreamName":
		return p.logStreamName
	case "Region":
		return p.region
	case "AutoCreateStream":
		return p.autoCreateStream
	case "StateFile":
		return p.stateFile
	}
	return "unknown-" + key
}

func (p *testFluentPlugin) Unregister(ctx unsafe.Pointer) {}
func (p *testFluentPlugin) GetRecord(dec *output.FLBDecoder) (int, interface{}, map[interface{}]interface{}) {
	if p.position < len(p.records) {
		r := p.records[p.position]
		p.position++
		return r.rc, r.ts, r.data
	}
	return -1, nil, nil
}
func (p *testFluentPlugin) NewDecoder(data unsafe.Pointer, length int) *output.FLBDecoder { return nil }
func (p *testFluentPlugin) Exit(code int)                                                 {}
func (p *testFluentPlugin) Put(timestamp time.Time, line string, sequenceToken string) (*cloudwatchlogs.PutLogEventsOutput, error) {
	data := ([]byte)(line)
	events := &events{data: data}
	p.events = append(p.events, events)
	return nil, nil
}

func (p *testFluentPlugin) CreateLogGroup(logGroupName string) error {
	return nil
}

func (p *testFluentPlugin) CreateLogStream(logGroupName, logStreamName string) error {
	return nil
}

func (p *testFluentPlugin) addrecord(rc int, ts interface{}, line map[interface{}]interface{}) {
	p.records = append(p.records, testrecord{rc: rc, ts: ts, data: line})
}

type stubProvider struct {
	creds   credentials.Value
	expired bool
	err     error
}

func (s *stubProvider) Retrieve() (credentials.Value, error) {
	s.expired = false
	s.creds.ProviderName = "stubProvider"
	return s.creds, s.err
}
func (s *stubProvider) IsExpired() bool {
	return s.expired
}

type testCloudwatchLogsCredential struct {
	credential string
}

func (c *testCloudwatchLogsCredential) GetCredentials(accessID, secretkey, credential string) (*credentials.Credentials, error) {
	creds := credentials.NewCredentials(&stubProvider{
		creds: credentials.Value{
			AccessKeyID:     "AKID",
			SecretAccessKey: "SECRET",
			SessionToken:    "",
		},
		expired: true,
	})

	return creds, nil
}

func TestPluginInitializationWithStaticCredentials(t *testing.T) {
	cloudwatchLogsCreds = &testCloudwatchLogsCredential{}
	_, err := getCloudWatchLogsConfig("exampleaccessID", "examplesecretkey", "", "examplegroup", "examplestream", "exampleregion", "", "")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	plugin = &testFluentPlugin{
		accessKeyID:      "exampleaccesskeyid",
		secretAccessKey:  "examplesecretaccesskey",
		logGroupName:     "examplegroup",
		logStreamName:    "examplestream",
		region:           "exampleregion",
		autoCreateStream: "true",
		stateFile:        "",
	}
	res := FLBPluginInit(nil)
	assert.Equal(t, output.FLB_OK, res)
}

func TestPluginInitializationWithSharedCredentials(t *testing.T) {
	cloudwatchLogsCreds = &testCloudwatchLogsCredential{}
	_, err := getCloudWatchLogsConfig("", "", "examplecredentials", "examplegroup", "examplestream", "exampleregion", "true", "")
	if err != nil {
		t.Fatalf("failed test %#v", err)
	}
	plugin = &testFluentPlugin{
		credential:       "examplecredentials",
		logGroupName:     "examplegroup",
		logStreamName:    "examplestream",
		region:           "exampleregion",
		autoCreateStream: "true",
		stateFile:        "",
	}
	res := FLBPluginInit(nil)
	assert.Equal(t, output.FLB_OK, res)
}

func TestPluginFlusher(t *testing.T) {
	testplugin := &testFluentPlugin{
		credential:       "examplecredentials",
		accessKeyID:      "exampleaccesskeyid",
		secretAccessKey:  "examplesecretaccesskey",
		logGroupName:     "examplegroup",
		logStreamName:    "examplestream",
		autoCreateStream: "true",
	}
	ts := time.Date(2019, time.March, 10, 10, 11, 12, 0, time.UTC)
	testrecords := map[interface{}]interface{}{
		"mykey": "myvalue",
	}
	testplugin.addrecord(0, output.FLBTime{Time: ts}, testrecords)
	testplugin.addrecord(0, uint64(ts.Unix()), testrecords)
	testplugin.addrecord(0, 0, testrecords)
	plugin = testplugin
	res := FLBPluginFlush(nil, 0, nil)
	assert.Equal(t, output.FLB_OK, res)
	assert.Len(t, testplugin.events, len(testplugin.records))
	var parsed map[string]interface{}
	json.Unmarshal(testplugin.events[0].data, &parsed)
	assert.Equal(t, testrecords["mykey"], parsed["mykey"])
	json.Unmarshal(testplugin.events[1].data, &parsed)
	json.Unmarshal(testplugin.events[2].data, &parsed)
}
