<a id="markdown-nexpose-scan-notifier" name="nexpose-scan-notifier"></a>
# Nexpose Scan Notifier
[![GoDoc](https://godoc.org/github.com/asecurityteam/nexpose-scan-notifier?status.svg)](https://godoc.org/github.com/asecurityteam/nexpose-scan-notifier)
[![Build Status](https://travis-ci.com/asecurityteam/nexpose-scan-notifier.png?branch=master)](https://travis-ci.com/asecurityteam/nexpose-scan-notifier)
[![codecov.io](https://codecov.io/github/asecurityteam/nexpose-scan-notifier/coverage.svg?branch=master)](https://codecov.io/github/asecurityteam/nexpose-scan-notifier?branch=master)

<https://github.com/asecurityteam/nexpose-scan-notifier>

<!-- TOC -->

- [Nexpose Scan Notifier](#nexpose-scan-notifier)
  - [Overview](#overview)
  - [Configuration](#configuration)
    - [Timestamp Storage](#timestamp-storage)
      - [DynamoDB](#dynamodb)
      - [Dependency Check](#dependencycheck)
  - [Status](#status)
  - [Contributing](#contributing)
    - [Building And Testing](#building-and-testing)
    - [Quality Gates](#quality-gates)
    - [License](#license)
    - [Contributing Agreement](#contributing-agreement)

<!-- /TOC -->

<a id="markdown-overview" name="overview"></a>
## Overview

Nexpose Scan Notifier is an API service which queries Nexpose and generates events for completed scans.

<a id="markdown-configuration" name="configuration"></a>
## Configuration

<a id="markdown-timestamp-storage" name="timestamp-storage"></a>
### Timestamp Storage

This project depends on a mechanism to persist and retrieve the timestamp of the last processed scan. This ensures that
successfully processed scans are not reprocessed, and any scans which are not successfully produced can be retried.

The current implementation of the timestamp storage interface is a DynamoDB table.

<a id="markdown-dynamodb" name="dynamodb"></a>
#### DynamoDB

This project stores the timestamp of the last processed scan in a simple DynamoDB table. The table uses a static
partition key (which uses "lastProcessed" as its default value), and successfully processed timestamps are upserted with
the key "timestamp". The table schema would look like the following:

```json
{
    TableName : "ScanTimestamp",
    KeySchema: [
        {
            AttributeName: "partitionkey",
            KeyType: "HASH", //Partition key
        }
    ],
    AttributeDefinitions: [
        {
            AttributeName: "partitionkey",
            AttributeType: "S"
        }
    ]
}
```

<a id="markdown-dependencycheck" name="dependencycheck"></a>
### Dependency Check
Depending on the user, this service or app can be composed of a bunch of sidecars. While one can check whether the configuration and
placement of these sidecars are configured correctly internally it might be useful to check whether environment variables point
to the correct external dependencies.

An obvious external dependency would be Nexpose itself. Consider configuring `DEPENDENCYCHECK_NEXPOSEENDPOINT` within `docker-compose.yaml`, that way
users can check whether they are able to connect to Nexpose with `/dependencycheck`(example in `gateway-incoming.yaml`).

<a id="markdown-status" name="status"></a>
## Status

This project is in incubation which means we are not yet operating this tool in production
and the interfaces are subject to change.

<a id="markdown-contributing" name="contributing"></a>
## Contributing

<a id="markdown-building-and-testing" name="building-and-testing"></a>
### Building And Testing

We publish a docker image called [SDCLI](https://github.com/asecurityteam/sdcli) that
bundles all of our build dependencies. It is used by the included Makefile to help make
building and testing a bit easier. The following actions are available through the Makefile:

-   make dep

    Install the project dependencies into a vendor directory

-   make lint

    Run our static analysis suite

-   make test

    Run unit tests and generate a coverage artifact

-   make integration

    Run integration tests and generate a coverage artifact

-   make coverage

    Report the combined coverage for unit and integration tests

-   make build

    Generate a local build of the project (if applicable)

-   make run

    Run a local instance of the project (if applicable)

-   make doc

    Generate the project code documentation and make it viewable
    locally.

<a id="markdown-quality-gates" name="quality-gates"></a>
### Quality Gates

Our build process will run the following checks before going green:

-   make lint
-   make test
-   make integration
-   make coverage (combined result must be 85% or above for the project)

Running these locally, will give early indicators of pass/fail.

<a id="markdown-license" name="license"></a>
### License

This project is licensed under Apache 2.0. See LICENSE.txt for details.

<a id="markdown-contributing-agreement" name="contributing-agreement"></a>
### Contributing Agreement

Atlassian requires signing a contributor's agreement before we can accept a
patch. If you are an individual you can fill out the
[individual CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=3f94fbdc-2fbe-46ac-b14c-5d152700ae5d).
If you are contributing on behalf of your company then please fill out the
[corporate CLA](https://na2.docusign.net/Member/PowerFormSigning.aspx?PowerFormId=e1c17c66-ca4d-4aab-a953-2c231af4a20b).
