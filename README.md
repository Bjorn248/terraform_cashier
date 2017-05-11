# terraform_cashier

[![Go Report Card](https://goreportcard.com/badge/github.com/Bjorn248/terraform_cashier)](https://goreportcard.com/report/github.com/Bjorn248/terraform_cashier)
[![Build Status](https://travis-ci.org/BjornTwitchBot/terraform_cashier.svg?branch=master)](https://travis-ci.org/BjornTwitchBot/terraform_cashier)
[![codecov](https://codecov.io/gh/BjornTwitchBot/terraform_cashier/branch/master/graph/badge.svg)](https://codecov.io/gh/BjornTwitchBot/terraform_cashier)

This uses https://github.com/Bjorn248/graphql_aws_pricing_api to get pricing data

Designed to analyze terraform template files and return a cost estimate of running the infrastructure, assuming AWS is the target cloud. Perhaps other clouds can be supported going forward?

This is very much in a prototype state right now. Any advice or assistance is appreciated.

## Plan File
This relies on terraform plan files generated using `terraform plan -out=<filename>`.
It is recommended that you plan against an empty state so that all of your resources
are present in the plan file.

## Environment Variables
Variable Name | Description
------------ | -------------
AWS_REGION | The Region for which you want to create a price estimation (e.g. `us-east-1`)
TERRAFORM_PLANFILE | Where cashier should find your terraform plan output.
RUNNING_HOURS | (Optional) The number of running hours normally used in a month for your resources, on average. Defaults to 730 assuming 24/7 operation.
