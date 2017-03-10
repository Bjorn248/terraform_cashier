package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl"
	"io/ioutil"
	"strings"
)

type GraphQLHTTPRequestBody struct {
	Query         string `json:"query"`
	Variables     string `json:"variables"`
	OperationName string `json:"operationName"`
}

var knownResourceTypes = map[string]string{
	"aws_instance": "%s: AmazonEC2(Location:\"US East (N. Virginia)\", TermType:\"OnDemand\", InstanceType:\"%s\", OS:\"Linux\", Tenancy:\"Shared\") {PricePerUnit Unit Currency}",
}

// const apiUrl = "https://fvaexi95f8.execute-api.us-east-1.amazonaws.com/Dev/graphql"

func main() {
	file, err := ioutil.ReadFile("./terraform_example.tf")
	if err != nil {
		fmt.Printf("Error reading file: '%s'", err)
	}

	var decodedOutput map[string]interface{}

	unmarshalErr := hcl.Unmarshal(file, &decodedOutput)
	if unmarshalErr != nil {
		fmt.Printf("Error decoding HCL: '%s'", unmarshalErr)
	}

	masterResourceMap := map[string]map[string]int{
		"aws_instance": map[string]int{"r3.xlarge": 0},
	}

	arrayOfResources := decodedOutput["resource"].([]map[string]interface{})

	for _, resource := range arrayOfResources {
		for key := range resource {
			fmt.Println("Key: ", key)

			switch key {
			case "aws_instance":
				fmt.Println("This is an AWS Instance")
				resourceKeys := resource[key].([]map[string]interface{})
				for resourceKey := range resourceKeys[0] {
					instanceType := resourceKeys[0][resourceKey].([]map[string]interface{})[0]["instance_type"].(string)
					fmt.Println("Instance Type: ", instanceType)
					countResource(masterResourceMap, key, instanceType)
				}
			default:
				fmt.Println("resource type not recognized")
			}
		}
	}
	graphQLQueryString := generateGraphQLQuery(masterResourceMap)
	fmt.Println(graphQLQueryString)
}

// This doesn't need a pointer, or to return anything, because maps in go are always passed by reference
func countResource(resourceMap map[string]map[string]int, resourceName string, resourceType string) {
	if count := resourceMap[resourceName][resourceType]; count == 0 {
		resourceMap[resourceName][resourceType] = 1
	} else {
		resourceMap[resourceName][resourceType] = count + 1
	}
}

func generateGraphQLQuery(masterResourceMap map[string]map[string]int) string {
	graphQLQueryString := ""
	requestBody := GraphQLHTTPRequestBody{
		Query:         "",
		Variables:     "",
		OperationName: "",
	}
	for resource := range masterResourceMap {
		if queryStringTemplate, ok := knownResourceTypes[resource]; ok {
			for resourceType, count := range masterResourceMap[resource] {
				if count > 0 {
					alias := strings.Replace(resourceType, ".", "_", -1)
					graphQLQueryString = graphQLQueryString + " " + fmt.Sprintf(queryStringTemplate, alias, resourceType)
				}
			}
		}
	}
	graphQLQueryString = "{" + graphQLQueryString + "}"
	requestBody.Query = graphQLQueryString
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		fmt.Println(err)
		return "error"
	}
	return string(requestBodyJSON)
}
