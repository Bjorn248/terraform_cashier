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

	mockTerraformResources := map[string]map[string]int{
		"aws_instance": map[string]int{
			"r3.xlarge": 3,
			"m4.large":  1,
			"r4.xlarge": 3,
			"m4.xlarge": 1,
		},
	}

	var resourceCostMap map[string]float32
	var err error

	resourceCostMap, err = calculateInfraCost(mockResponse, mockTerraformResources)
	if resourceCostMap["aws_instance"] != 2.3792 {
		t.Error("Expected 2.3792, got ", resourceCostMap["aws_instance"])
	}
	if err != nil {
		t.Error("Something went wrong", err)
	}
}

func TestProcessTerraformFile(t *testing.T) {

	mockResourceMap := map[string]map[string]int{
		"aws_instance": {"r4.xlarge": 0},
	}

	processTerraformFile(mockResourceMap, "terraform_example.tf")
	processTerraformFile(mockResourceMap, "terraform_example_2.tf")

	if mockResourceMap["aws_instance"]["r3.xlarge"] != 3 ||
		mockResourceMap["aws_instance"]["m4.large"] != 1 ||
		mockResourceMap["aws_instance"]["r4.xlarge"] != 3 ||
		mockResourceMap["aws_instance"]["m4.xlarge"] != 1 {
		t.Error("Did not get expected results", mockResourceMap)
	}
}

func TestGenerateGraphQLQuery(t *testing.T) {

	mockTerraformResources := map[string]map[string]int{
		"aws_instance": map[string]int{
			"r3.xlarge": 3,
			"m4.large":  1,
			"r4.xlarge": 3,
			"m4.xlarge": 1,
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

	mockTerraformResources := map[string]map[string]int{
		"aws_instance": map[string]int{
			"r3.xlarge": 3,
			"m4.large":  1,
			"r4.xlarge": 3,
			"m4.xlarge": 1,
		},
	}

	countResource(mockTerraformResources, "aws_instance", "t2.small")
	countResource(mockTerraformResources, "aws_instance", "t2.medium")
	countResource(mockTerraformResources, "aws_instance", "m4.xlarge")
	if mockTerraformResources["aws_instance"]["r3.xlarge"] != 3 ||
		mockTerraformResources["aws_instance"]["m4.large"] != 1 ||
		mockTerraformResources["aws_instance"]["r4.xlarge"] != 3 ||
		mockTerraformResources["aws_instance"]["t2.small"] != 1 ||
		mockTerraformResources["aws_instance"]["t2.medium"] != 1 ||
		mockTerraformResources["aws_instance"]["m4.xlarge"] != 2 {
		t.Error("Did not get expected results", mockTerraformResources)
	}
}
