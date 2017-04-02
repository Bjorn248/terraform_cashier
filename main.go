package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hcl"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type graphQLHTTPRequestBody struct {
	Query         string `json:"query"`
	Variables     string `json:"variables"`
	OperationName string `json:"operationName"`
}

type graphQLHTTPResponseBody struct {
	Data map[string][]graphQLResponseData `json:"data"`
}

type graphQLResponseData struct {
	PricePerUnit string `json:"PricePerUnit"`
	Unit         string `json:"Unit"`
	Currency     string `json:"Currency"`
}

var knownResourceTypes = map[string]string{
	// Using query aliases to get the pricing data for different types of instances at the same time
	"aws_instance": "%s: AmazonEC2(Location:\"%s\", TermType:\"OnDemand\", InstanceType:\"%s\", OS:\"Linux\", Tenancy:\"Shared\") {PricePerUnit Unit Currency}",
}

var resourceTypesToFriendlyNames = map[string]string{
	"aws_instance": "EC2",
}

var regionMap = map[string]string{
	"us-gov-west-1":  "AWS GovCloud (US)",
	"us-east-1":      "US East (N. Virginia)",
	"us-east-2":      "US East (Ohio)",
	"us-west-1":      "US West (N. California)",
	"us-west-2":      "US West (Oregon)",
	"ca-central-1":   "Canada (Central)",
	"eu-west-1":      "EU (Ireland)",
	"eu-central-1":   "EU (Frankfurt)",
	"eu-west-2":      "EU (London)",
	"ap-northeast-1": "Asia Pacific (Tokyo)",
	"ap-northeast-2": "Asia Pacific (Seoul)",
	"ap-southeast-1": "Asia Pacific (Singapore)",
	"ap-southeast-2": "Asia Pacific (Sydney)",
	"ap-south-1":     "Asia Pacific (Mumbai)",
	"sa-east-1":      "South America (Sao Paulo)",
}

// See https://github.com/Bjorn248/graphql_aws_pricing_api for the code of this API
const apiUrl = "https://fvaexi95f8.execute-api.us-east-1.amazonaws.com/Dev/graphql"

func main() {

	if os.Getenv("AWS_REGION") == "" {
		log.Fatal("AWS_REGION not set")
	}

	// We want all the .tf files in the current directory
	pattern := "*.tf"

	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("Error reading files: '%s'", err)
	}

	masterResourceMap := map[string]map[string]int{
		"aws_instance": {"r3.xlarge": 0},
	}

	for _, filePath := range matches {
		processTerraformFile(masterResourceMap, filePath)
	}

	graphQLQueryString, err := generateGraphQLQuery(masterResourceMap)
	if err != nil {
		fmt.Printf("Error generating GraphQL Query: '%s'", err)
	}
	// fmt.Println("GraphQL Query String: ", graphQLQueryString)

	// We want a high timeout because the lambda function
	// needs at least 1 request to warm up. The first request
	// always takes a long time.
	timeout := time.Duration(40 * time.Second)

	client := http.Client{
		Timeout: timeout,
	}

	fmt.Println("Calling GraphQL Pricing API...")

	resp, err := client.Post(apiUrl, "application/json", bytes.NewBuffer([]byte(graphQLQueryString)))
	if err != nil {
		fmt.Printf("Error making request to Pricing API: '%s'", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: '%s'", err)
	}
	var response graphQLHTTPResponseBody

	unmarshalErr := json.Unmarshal(body, &response)
	if unmarshalErr != nil {
		log.Fatal("Error Unmarshalling Response Body", unmarshalErr)
	}

	resourceCostMap, err := calculateInfraCost(response, masterResourceMap)
	if err != nil {
		log.Fatal("Error generating resourceCostMap", err)
	}

	for resource, hourlyCost := range resourceCostMap {
		fmt.Println("")
		fmt.Println("Cost of", resourceTypesToFriendlyNames[resource])
		fmt.Printf("Hourly: $%v\nMonthly: $%v\nNote: Monthly Cost based on 730 hours on average being in a month\n", hourlyCost, hourlyCost*float32(730))
	}
}

// This function takes the pricing data and uses it to calculate the infrastructure cost by looking at the
// map of terraform resources. Basically, we're just iterating over some maps here...
func calculateInfraCost(pricingData graphQLHTTPResponseBody, terraformResources map[string]map[string]int) (map[string]float32, error) {
	resourceCostMap := map[string]float32{
		"aws_instance": 0.00,
	}

	for resourceName, resourceTypes := range terraformResources {
		resourceCostMap[resourceName] = 0.00
		for resourceType, count := range resourceTypes {
			alias := strings.Replace(resourceType, ".", "_", -1)
			oldValue := resourceCostMap[resourceName]
			price, err := strconv.ParseFloat(pricingData.Data[alias][0].PricePerUnit, 32)
			if err != nil {
				return resourceCostMap, err
			}
			resourceCostMap[resourceName] = oldValue + (float32(price) * float32(count))
		}
	}
	return resourceCostMap, nil
}

// This takes a terraform file and adds it to the global resource map used to shape the GraphQL query
func processTerraformFile(masterResourceMap map[string]map[string]int, filePath string) {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: '%s'", err)
	}

	var decodedOutput map[string]interface{}

	unmarshalErr := hcl.Unmarshal(file, &decodedOutput)
	if unmarshalErr != nil {
		fmt.Printf("Error decoding HCL: '%s'", unmarshalErr)
	}

	arrayOfResources := decodedOutput["resource"].([]map[string]interface{})

	for _, resource := range arrayOfResources {
		for key := range resource {
			switch key {
			case "aws_instance":
				resourceKeys := resource[key].([]map[string]interface{})
				for resourceKey := range resourceKeys[0] {
					instanceType := resourceKeys[0][resourceKey].([]map[string]interface{})[0]["instance_type"].(string)
					countResource(masterResourceMap, key, instanceType)
				}
			default:
				fmt.Println("resource type not recognized: ", key)
			}
		}
	}
}

// This doesn't need a pointer, or to return anything, because maps in go are always passed by reference
func countResource(resourceMap map[string]map[string]int, resourceName string, resourceType string) {
	if count := resourceMap[resourceName][resourceType]; count == 0 {
		resourceMap[resourceName][resourceType] = 1
	} else {
		resourceMap[resourceName][resourceType] = count + 1
	}
}

func generateGraphQLQuery(masterResourceMap map[string]map[string]int) (string, error) {
	graphQLQueryString := ""
	requestBody := graphQLHTTPRequestBody{
		Query:         "",
		Variables:     "",
		OperationName: "",
	}
	for resource := range masterResourceMap {
		if queryStringTemplate, ok := knownResourceTypes[resource]; ok {
			for resourceType, count := range masterResourceMap[resource] {
				if count > 0 {
					region := regionMap[os.Getenv("AWS_REGION")]
					alias := strings.Replace(resourceType, ".", "_", -1)
					graphQLQueryString = graphQLQueryString + " " + fmt.Sprintf(queryStringTemplate, alias, region, resourceType)
				}
			}
		}
	}
	graphQLQueryString = "{" + graphQLQueryString + "}"
	requestBody.Query = graphQLQueryString
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	return string(requestBodyJSON), nil
}
