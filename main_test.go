package main

import (
	"encoding/json"
	"testing"
)

func TestCalculateInfraCost(t *testing.T) {
	mockResponseData := `{
	"data": {
		"r3_xlarge_Dedicated": [
			{
				"PricePerUnit": "0.408",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		],
		"r3_xlarge_Shared": [
			{
				"PricePerUnit": "0.371",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		],
		"m4_large_Shared": [
			{
				"PricePerUnit": "0.126",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		],
		"r4_xlarge_Shared": [
			{
				"PricePerUnit": "0.2964",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		],
		"m4_xlarge_Shared": [
			{
				"PricePerUnit": "0.251",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		],
		"db_t2_large_mysql_Single_AZ": [
			{
				"PricePerUnit": "0.178",
				"Unit": "Hrs",
				"Currency": "USD"
			}
		],
		"db_t2_large_mysql_Multi_AZ": [
			{
				"PricePerUnit": "0.356",
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
				"r3.xlarge,Shared":    3,
				"r3.xlarge,Dedicated": 1,
				"m4.large,Shared":     1,
				"r4.xlarge,Shared":    3,
				"m4.xlarge,Shared":    1,
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
	if resourceCostMapArray[0].Total != 2.7872 {
		t.Error("Expected 2.7872, got ", resourceCostMapArray[0].Total)
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
				"r4.xlarge,Shared": 0,
			},
			"aws_db_instance": {
				"db.t2.large,mysql,Single-AZ": 0,
			},
		},
	}

	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "test/terraform_example.tf")
	if err != nil {
		t.Error("error processing files", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "test/terraform_example_2.tf")
	if err != nil {
		t.Error("error processing files", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "test/variables.tf")
	if err.Error() != "Could not find resources in test/variables.tf" {
		t.Error("Expected a different error string", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "test/does_not_exist.tf")
	if err.Error() != "open test/does_not_exist.tf: no such file or directory" {
		t.Error("Expected a different error string", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "test/bad_formatting.tf")
	if err.Error() != "At 2:17: literal not terminated" {
		t.Error("Expected a different error string", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "test/rds.tf")
	if err != nil {
		t.Error("error processing test/rds.tf", err)
	}
	mockTerraformResources, err = processTerraformFile(mockTerraformResources, "test/ec2_dedicated.tf")
	if err != nil {
		t.Error("error processing test/ec2_dedicated.tf", err)
	}

	if mockTerraformResources.Resources["aws_instance"]["r3.xlarge,Shared"] != 3 ||
		mockTerraformResources.Resources["aws_instance"]["m4.large,Shared"] != 1 ||
		mockTerraformResources.Resources["aws_instance"]["r4.xlarge,Shared"] != 3 ||
		mockTerraformResources.Resources["aws_instance"]["m4.xlarge,Shared"] != 1 ||
		mockTerraformResources.Resources["aws_instance"]["r3.xlarge,Dedicated"] != 1 ||
		mockTerraformResources.Resources["aws_db_instance"]["db.t2.large,mysql,Multi-AZ"] != 1 ||
		mockTerraformResources.Resources["aws_db_instance"]["db.t2.large,mysql,Single-AZ"] != 2 {
		t.Error("Didn't not get expected results", mockTerraformResources)
	}
}

func TestGenerateGraphQLQuery(t *testing.T) {

	mockTerraformResources := resourceMap{
		Resources: map[string]map[string]int{
			"aws_instance": {
				"r3.xlarge,Shared":    3,
				"r3.xlarge,Dedicated": 1,
				"m4.large,Shared":     1,
				"r4.xlarge,Shared":    3,
				"m4.xlarge,Shared":    1,
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
				"r3.xlarge,Shared":    3,
				"r3.xlarge,Dedicated": 1,
				"m4.large,Shared":     1,
				"r4.xlarge,Shared":    3,
				"m4.xlarge,Shared":    1,
			},
			"aws_db_instance": {
				"db.r4.xlarge,mysql,Single-AZ": 3,
				"db.t2.large,mysql,Multi-AZ":   1,
			},
		},
	}

	mockTerraformResources = countResource(mockTerraformResources, "aws_instance", "t2.small,Shared")
	mockTerraformResources = countResource(mockTerraformResources, "aws_instance", "t2.medium,Shared")
	mockTerraformResources = countResource(mockTerraformResources, "aws_instance", "m4.xlarge,Shared")
	if mockTerraformResources.Resources["aws_instance"]["r3.xlarge,Shared"] != 3 ||
		mockTerraformResources.Resources["aws_instance"]["m4.large,Shared"] != 1 ||
		mockTerraformResources.Resources["aws_instance"]["r4.xlarge,Shared"] != 3 ||
		mockTerraformResources.Resources["aws_instance"]["t2.small,Shared"] != 1 ||
		mockTerraformResources.Resources["aws_instance"]["t2.medium,Shared"] != 1 ||
		mockTerraformResources.Resources["aws_instance"]["m4.xlarge,Shared"] != 2 {
		t.Error("Did not get expected results", mockTerraformResources)
	}
}
