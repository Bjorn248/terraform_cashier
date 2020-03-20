# terraform_cashier

[![Go Report Card](https://goreportcard.com/badge/github.com/Bjorn248/terraform_cashier)](https://goreportcard.com/report/github.com/Bjorn248/terraform_cashier)
[![Build Status](https://travis-ci.org/BjornTwitchBot/terraform_cashier.svg?branch=master)](https://travis-ci.org/BjornTwitchBot/terraform_cashier)
[![codecov](https://codecov.io/gh/BjornTwitchBot/terraform_cashier/branch/master/graph/badge.svg)](https://codecov.io/gh/BjornTwitchBot/terraform_cashier)
[![Maintainability](https://api.codeclimate.com/v1/badges/f66843242e5aeda56ba8/maintainability)](https://codeclimate.com/github/BjornTwitchBot/terraform_cashier/maintainability)

This uses [https://github.com/Bjorn248/graphql_aws_pricing_api](https://github.com/Bjorn248/graphql_aws_pricing_api)
to get pricing data

Designed to analyze terraform template files and return a cost estimate of running the
infrastructure, assuming AWS is the target cloud. Perhaps other clouds can be supported going forward?

**NOTE**: This only calculates the cost of EC2 and RDS resources right now. To add support for
more resources, open a PR or leave a comment. I'm always looking for feedback.

This is very much in a prototype state right now. Any advice or assistance is appreciated.

## Plan File
This relies on terraform plan files generated using `terraform plan -out=<filename>`.
It is recommended that you plan against an empty state so that all of your resources
are present in the plan file.

### Empty/Blank state plan
If you are using remote state, you will need to disable this. I accomplish this by renaming the
file that defines my remote state, appending the extension `.off`. Then, I have to run
`terraform init` again to initialize a blank local state. Then I can generate a plan that contains
all my resources using the following command `terraform plan -out=<filename> -refresh=false`.

## Environment Variables
Variable Name      | Description
------------       | -------------
AWS_REGION         | The Region for which you want to create a price estimation (e.g. `us-east-1`)
TERRAFORM_PLANFILE | Where cashier should find your terraform plan output.
RUNNING_HOURS      | (Optional) The number of running hours normally used in a month for your resources, on average. Defaults to 730 assuming 24/7 operation.
GRAPHQL_API_URL    | (Optional) Change the API URL that cashier uses. Defaults to live API Gateway URL.
PRINT_VERSION      | (Optional) If `true` will print current version of cashier and exit

## Installation and usage
Simply download the latest release from the releases page here:
[https://github.com/Bjorn248/terraform_cashier/releases](https://github.com/Bjorn248/terraform_cashier/releases)
Make sure that the binary is set executable
Set your environment variables appropriately
```
export TERRAFORM_PLANFILE="./terraform.plan"
export AWS_REGION="us-west-1"
```
And then run the `cashier` binary

If you wish to see the current version of cashier, use the following environment variable
```
PRINT_VERSION=true
```


## Docker 
The Docker image is not published on Docker Hub or any other registry provider, although you can
build and use it locally.

*Build*
```
$ docker build -t local/cashier:0.0.1 .
```

*Run*
```
$ docker run -v $(pwd):/data local/cashier:0.0.1
```

*Run using custom ENV_VARs*
```
$ docker run --env AWS_REGION=us-east-1 --env TERRAFORM_PLANFILE=/data/terraform.plan -v $(pwd):/data local/cashier:0.0.1
```

## Local Development
This project uses Go Modules for dependency management. `go build` should just work if your Go version
supports modules.
