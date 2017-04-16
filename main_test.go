package main

import (
	"encoding/json"
	"testing"
)

func TestCalculateInfraCost(t *testing.T) {
	mockResponseData := `{
	"data": {
		"r3_xlarge": [
			{
				"PricePerUnit": "0.371",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		],
		"m4_large": [
			{
				"PricePerUnit": "0.126",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		],
		"r4_xlarge": [
			{
				"PricePerUnit": "0.2964",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		],
		"m4_xlarge": [
			{
				"PricePerUnit": "0.251",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		]
	}
}`

	var mockResponse graphQLHTTPResponseBody
	unmarshalErr := json.Unmarshal([]byte(mockResponseData), &mockResponse)
	if unmarshalErr != nil {
		t.Error("Error Unmarshalling Mock Response Body", unmarshalErr)
	}

	mockTerraformResources := resourceMap{
		Resources: map[string]map[string]int{
			"aws_instance": {
				"r3.xlarge": 3,
				"m4.large":  1,
				"r4.xlarge": 3,
				"m4.xlarge": 1,
			},
			"aws_db_instance": {
				"db.r4.xlarge,mysql,Single-AZ": 3,
				"db.t2.large,mysql,Multi-AZ":   1,
			},
		},
	}

	var resourceCostMapArray []resourceCostMap
	var err error

	resourceCostMapArray, err = calculateInfraCost(mockResponse, mockTerraformResources)
	if resourceCostMapArray[0].Total != 2.3792 {
		t.Error("Expected 2.3792, got ", resourceCostMapArray[0].Total)
	}
	if err != nil {
		t.Error("Something went wrong", err)
	}
}

func TestProcessTerraformFile(t *testing.T) {

	var err error

	mockTerraformResources := resourceMap{
		Resources: map[string]map[string]int{
			"aws_instance": {
				"r4.xlarge": 0,
			},
			"aws_db_instance": {
				"db.t2.large,mysql,Single-AZ": 0,
			},
		},
	}

	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "terraform_example.tf")
	if err != nil {
		t.Error("error processing files", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "terraform_example_2.tf")
	if err != nil {
		t.Error("error processing files", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "variables.tf")
	if err.Error() != "Could not find resources in variables.tf" {
		t.Error("Expected a different error string", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "does_not_exist.tf")
	if err.Error() != "open does_not_exist.tf: no such file or directory" {
		t.Error("Expected a different error string", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "bad_formatting.tf")
	if err.Error() != "At 2:17: literal not terminated" {
		t.Error("Expected a different error string", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "rds.tf")
	if err != nil {
		t.Error("error processing rds.tf", err)
	}

	if mockTerraformResources.Resources["aws_instance"]["r3.xlarge"] != 3 ||
		mockTerraformResources.Resources["aws_instance"]["m4.large"] != 1 ||
		mockTerraformResources.Resources["aws_instance"]["r4.xlarge"] != 3 ||
		mockTerraformResources.Resources["aws_instance"]["m4.xlarge"] != 1 ||
		mockTerraformResources.Resources["aws_db_instance"]["db.t2.large,mysql,Multi-AZ"] != 1 ||
		mockTerraformResources.Resources["aws_db_instance"]["db.t2.large,mysql,Single-AZ"] != 2 {
		t.Error("Didn't not get expected results", mockTerraformResources)
	}
}

func TestGenerateGraphQLQuery(t *testing.T) {

	mockTerraformResources := resourceMap{
		Resources: map[string]map[string]int{
			"aws_instance": {
				"r3.xlarge": 3,
				"m4.large":  1,
				"r4.xlarge": 3,
				"m4.xlarge": 1,
			},
			"aws_db_instance": {
				"db.r4.xlarge,mysql,Single-AZ": 3,
				"db.t2.large,mysql,Multi-AZ":   1,
			},
		},
	}

	graphQLQueryString, err := generateGraphQLQuery(mockTerraformResources)
	if err != nil {
		t.Error("Something went wrong generating the query string", err)
	}

	var requestBody graphQLHTTPRequestBody
	unmarshalErr := json.Unmarshal([]byte(graphQLQueryString), &requestBody)
	if unmarshalErr != nil {
		t.Error("Did not generate expected query string", unmarshalErr)
	}
}

func TestCountResource(t *testing.T) {

	mockTerraformResources := resourceMap{
		Resources: map[string]map[string]int{
			"aws_instance": {
				"r3.xlarge": 3,
				"m4.large":  1,
				"r4.xlarge": 3,
				"m4.xlarge": 1,
			},
			"aws_db_instance": {
				"db.r4.xlarge,mysql,Single-AZ": 3,
				"db.t2.large,mysql,Multi-AZ":   1,
			},
		},
	}

	mockTerraformResources = countResource(mockTerraformResources, "aws_instance", "t2.small")
	mockTerraformResources = countResource(mockTerraformResources, "aws_instance", "t2.medium")
	mockTerraformResources = countResource(mockTerraformResources, "aws_instance", "m4.xlarge")
	if mockTerraformResources.Resources["aws_instance"]["r3.xlarge"] != 3 ||
		mockTerraformResources.Resources["aws_instance"]["m4.large"] != 1 ||
		mockTerraformResources.Resources["aws_instance"]["r4.xlarge"] != 3 ||
		mockTerraformResources.Resources["aws_instance"]["t2.small"] != 1 ||
		mockTerraformResources.Resources["aws_instance"]["t2.medium"] != 1 ||
		mockTerraformResources.Resources["aws_instance"]["m4.xlarge"] != 2 {
		t.Error("Did not get expected results", mockTerraformResources)
	}
}
