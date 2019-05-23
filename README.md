# fluent-bit cloudwatch-logs output plugin

[![Build Status](https://travis-ci.org/cosmo0920/fluent-bit-go-cloudwatch-logs.svg?branch=master)](https://travis-ci.org/cosmo0920/fluent-bit-go-cloudwatch-logs)

This plugin works with fluent-bit's go plugin interface. You can use fluent-bit-go-cloudwatch-logs to ship logs into AWS CloudWatch.

The configuration typically looks like:

```graphviz
fluent-bit --> AWS CloudWatch
```

# Usage

```bash
$ fluent-bit -e /path/to/built/out_cloudwatch_logs.so -c fluent-bit.conf
```

# Prerequisites

* Go 1.11+
* gcc (for cgo)

## Building

```bash
$ make
```

### Configuration Options

| Key               | Description                     | Default value |  Note                           |
|-------------------|---------------------------------|---------------|---------------------------------|
| Credential        | URI of AWS shared credential    | `""`          |(See [Credentials](#credentials))|
| AccessKeyID       | Access key ID of AWS            | `""`          |(See [Credentials](#credentials))|
| SecretAccessKey   | Secret access key ID of AWS     | `""`          |(See [Credentials](#credentials))|
| LogGroupName      | logGroup name of CloudWatch     | `-`           | Mandatory parameter             |
| LogStreamName     | logStream name of CloudWatch    | `-`           | Mandatory parameter             |
| Region            | Region of CloudWatch            | `-`           | Mandatory parameter             |
| AutoCreateStream  | Use auto create stream feature? | `true`        | Optional parameter              |
| StateFile         | filepath for saving state       | `""`          | Optional parameter              |

Example:

add this section to fluent-bit.conf

```properties
[Output]
    Name cloudwatch_logs
    Match *
    # Credential    /path/to/sharedcredentialfile
    AccessKeyID     yourawsaccesskeyid
    SecretAccessKey yourawssecretaccesskey
    LogGroupName    yourloggroupname
    LogStreamName   yourslogstreamname
    Region us-east-1
    # AutoCreateStream false # default: true
    # StateFile     yourstatefile
```

fluent-bit-go-cloudwatch-logs supports the following credentials. Users must specify one of them:

## Credentials

Specifying credentials is **required**.

This plugin supports the following credentials:

### Shared Credentials

Create the following file which includes credentials:

```ini
[default]
aws_access_key_id = YOUR_AWS_ACCESS_KEY_ID
aws_secret_access_key = YOUR_AWS_SECRET_ACCESS_KEY
```

And specify the following parameter in fluent-bit configuration:

```ini
Credential    /path/to/sharedcredentialfile
```

### Static Credentials

Specify the following parameters in fluent-bit configuration:

```ini
AccessKeyID     yourawsaccesskeyid
SecretAccessKey yourawssecretaccesskey
```

### Environment Credentials

Specify `AWS_ACCESS_KEY` and `AWS_SECRET_KEY` as environment variables.

## Useful links

* [fluent-bit-go](https://github.com/fluent/fluent-bit-go)
