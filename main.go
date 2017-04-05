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

type resourceCostMap struct {
	Resources map[string]float32
	Name      string
	Total     float32
}

type resourceMap struct {
	Resources map[string]map[string]int
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
	// notest

	if os.Getenv("AWS_REGION") == "" {
		log.Fatal("AWS_REGION not set")
	}

	pattern := ""
	if os.Getenv("GLOB_PATTERN") == "" {
		pattern = "*.tf"
	} else {
		pattern = os.Getenv("GLOB_PATTERN")
	}

	matches, err := filepath.Glob(pattern)
	if err != nil {
		fmt.Printf("Error reading files: '%s'", err)
	}

	masterResourceMap := resourceMap{
		Resources: map[string]map[string]int{"aws_instance": {"r4.xlarge": 0}},
	}

	for _, filePath := range matches {
		masterResourceMap = processTerraformFile(masterResourceMap, filePath)
	}

	graphQLQueryString, err := generateGraphQLQuery(masterResourceMap)
	if err != nil {
		fmt.Printf("Error generating GraphQL Query: '%s'", err)
	}
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

	resourceCostMapArray, err := calculateInfraCost(response, masterResourceMap)
	if err != nil {
		log.Fatal("Error generating []resourceCostMap", err)
	}

	var runningHours uint64
	if os.Getenv("RUNNING_HOURS") == "" {
		runningHours = 730
	} else {
		runningHours, err = strconv.ParseUint(os.Getenv("RUNNING_HOURS"), 10, 16)
		if err != nil {
			log.Fatal("Error parsing int from RUNNING_HOURS environment variable", err)
		}
	}

	for _, resourceCostMap := range resourceCostMapArray {
		fmt.Println("")
		fmt.Println("Cost of", resourceTypesToFriendlyNames[resourceCostMap.Name])
		fmt.Println("Breakdown by type:")
		for resourceType, cost := range resourceCostMap.Resources {
			if resourceType != "" {
				fmt.Printf("%v: $%v\n", resourceType, cost)
			}
		}
		fmt.Printf("Total Hourly: $%v\nTotal Monthly: $%v\nNote: Monthly cost based on %v runtime hours per month\n", resourceCostMap.Total, resourceCostMap.Total*float32(runningHours), runningHours)
		// fmt.Printf("Hourly: $%v\nMonthly: $%v\nNote: Monthly cost based on %v runtime hours per month\n", hourlyCost, hourlyCost*float32(runningHours), runningHours)
	}
}

// This function takes the pricing data and uses it to calculate the infrastructure cost by looking at the
// map of terraform resources. Basically, we're just iterating over some maps here...
func calculateInfraCost(pricingData graphQLHTTPResponseBody, terraformResources resourceMap) ([]resourceCostMap, error) {
	var returnArray []resourceCostMap

	for resourceName, resourceTypes := range terraformResources.Resources {
		var resourceSpecificCostMap resourceCostMap
		resourceSpecificCostMap.Name = resourceName
		resourceSpecificCostMap.Resources = map[string]float32{"": 0.00}
		for resourceType, count := range resourceTypes {
			alias := strings.Replace(resourceType, ".", "_", -1)
			var price float64
			var err error
			for _, element := range pricingData.Data[alias] {
				price, err = strconv.ParseFloat(element.PricePerUnit, 32)
				if err != nil {
					return []resourceCostMap{}, err
				}
			}
			resourceSpecificCostMap.Resources[resourceType] = (float32(price) * float32(count))
		}
		var runningTotalCost float32
		for _, cost := range resourceSpecificCostMap.Resources {
			runningTotalCost = resourceSpecificCostMap.Total
			resourceSpecificCostMap.Total = runningTotalCost + cost
		}
		returnArray = append(returnArray, resourceSpecificCostMap)
	}
	return returnArray, nil
}

// This takes a terraform file and adds it to the global resource map used to shape the GraphQL query
func processTerraformFile(masterResourceMap resourceMap, filePath string) resourceMap {
	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Printf("Error reading file: '%s'", err) // notest
	}

	var decodedOutput map[string]interface{}

	unmarshalErr := hcl.Unmarshal(file, &decodedOutput)
	if unmarshalErr != nil {
		fmt.Printf("Error decoding HCL: '%s'", unmarshalErr) // notest
	}

	arrayOfResources, success := decodedOutput["resource"].([]map[string]interface{})

	if success == true {
		for _, resource := range arrayOfResources {
			for key := range resource {
				switch key {
				case "aws_instance":
					resourceKeys, resourceKeysSuccess := resource[key].([]map[string]interface{})
					if resourceKeysSuccess == true {
						for resourceKey := range resourceKeys[0] {
							instanceType, instanceTypeSuccess := resourceKeys[0][resourceKey].([]map[string]interface{})[0]["instance_type"].(string)
							if instanceTypeSuccess == true {
								masterResourceMap = countResource(masterResourceMap, key, instanceType)
							}
						}
					}
				default:
					fmt.Println("resource type not recognized: ", key)
				}
			}
		}
	}
	return masterResourceMap
}

func countResource(masterResourceMap resourceMap, resourceName string, resourceType string) resourceMap {
	if count := masterResourceMap.Resources[resourceName][resourceType]; count == 0 {
		masterResourceMap.Resources[resourceName][resourceType] = 1
	} else {
		masterResourceMap.Resources[resourceName][resourceType] = count + 1
	}
	return masterResourceMap
}

func generateGraphQLQuery(masterResourceMap resourceMap) (string, error) {
	graphQLQueryString := ""
	requestBody := graphQLHTTPRequestBody{
		Query:         "",
		Variables:     "",
		OperationName: "",
	}
	for resource := range masterResourceMap.Resources {
		if queryStringTemplate, ok := knownResourceTypes[resource]; ok {
			for resourceType, count := range masterResourceMap.Resources[resource] {
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
