package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform/plans/planfile"
	"github.com/zclconf/go-cty/cty"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	"aws_instance":    "%s: AmazonEC2(Location:\"%s\", TermType:\"%s\", InstanceType:\"%s\", OS:\"Linux\", PreInstalledSW:\"NA\", CapacityStatus:\"Used\", Tenancy:\"%s\") {PricePerUnit Unit Currency}",
	"aws_db_instance": "%s: AmazonRDS(Location:\"%s\", TermType:\"%s\", InstanceType:\"%s\", DeploymentOption:\"%s\", DatabaseEngine:\"%s\") {PricePerUnit Unit Currency}",
}

var resourceTypesToFriendlyNames = map[string]string{
	"aws_instance":    "EC2",
	"aws_db_instance": "RDS",
}

var regionMap = map[string]string{
	"us-gov-west-1":  "AWS GovCloud (US)",
	"us-gov-east-1":  "AWS GovCloud (US-East)",
	"us-east-1":      "US East (N. Virginia)",
	"us-east-2":      "US East (Ohio)",
	"us-west-1":      "US West (N. California)",
	"us-west-2":      "US West (Oregon)",
	"ca-central-1":   "Canada (Central)",
	"cn-north-1":     "China (Beijing)",
	"cn-northwest-1": "China (Ningxia)",
	"eu-central-1":   "EU (Frankfurt)",
	"eu-west-1":      "EU (Ireland)",
	"eu-west-2":      "EU (London)",
	"eu-west-3":      "EU (Paris)",
	"eu-north-1":     "EU (Stockholm)",
	"ap-northeast-1": "Asia Pacific (Tokyo)",
	"ap-northeast-2": "Asia Pacific (Seoul)",
	"ap-northeast-3": "Asia Pacific (Osaka-Local)",
	"ap-southeast-1": "Asia Pacific (Singapore)",
	"ap-southeast-2": "Asia Pacific (Sydney)",
	"ap-south-1":     "Asia Pacific (Mumbai)",
	"sa-east-1":      "South America (Sao Paulo)",
}

// See https://github.com/Bjorn248/graphql_aws_pricing_api for the code of this API
const apiUrl = "https://fvaexi95f8.execute-api.us-east-1.amazonaws.com/Dev/graphql"

// Should match the git tagged release
const version = "0.6"

func main() {
	// notest

	if os.Getenv("PRINT_VERSION") == "true" {
		fmt.Println("Terraform Cashier")
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}

	if os.Getenv("AWS_REGION") == "" {
		log.Fatal("AWS_REGION not set")
	}

	var terraformPlanFile string
	if os.Getenv("TERRAFORM_PLANFILE") == "" {
		log.Fatal("TERRAFORM_PLANFILE not set")
	} else {
		terraformPlanFile = os.Getenv("TERRAFORM_PLANFILE")
	}

	masterResourceMap := resourceMap{
		Resources: map[string]map[string]int{
			"aws_instance":    {"r4.xlarge,Shared": 0},
			"aws_db_instance": {"db.r4.xlarge,mysql,Single-AZ": 0},
		},
	}

	var err error

	masterResourceMap, err = processTerraformPlan(masterResourceMap, terraformPlanFile)
	if err != nil {
		fmt.Printf("Error processing terraform plan: '%s'\n", err)
	}

	graphQLQueryString, err := generateGraphQLQuery(masterResourceMap)
	if err != nil {
		fmt.Printf("Error generating GraphQL Query: '%s'\n", err)
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

	var totalMonthlyCost float32

	for _, resourceCostMap := range resourceCostMapArray {
		fmt.Println("")
		fmt.Println("Cost of", resourceTypesToFriendlyNames[resourceCostMap.Name])
		fmt.Println("Breakdown by type:")
		for resourceType, cost := range resourceCostMap.Resources {
			if resourceType != "" && cost != 0.00 {
				fmt.Printf("%v (%v): $%v\n", resourceType, masterResourceMap.Resources[resourceCostMap.Name][resourceType], cost)
			}
		}
		fmt.Printf("Total Hourly: $%v\nTotal Monthly: $%v\nNote: Monthly cost based on %v runtime hours per month\n", resourceCostMap.Total, resourceCostMap.Total*float32(runningHours), runningHours)

		totalMonthlyCost = totalMonthlyCost + resourceCostMap.Total*float32(runningHours)
	}
	fmt.Println("")
	fmt.Printf("Total Monthly Cost of All Services: %v\n", totalMonthlyCost)
}

// This function takes the pricing data and uses it to calculate the infrastructure cost by looking at the
// map of terraform resources. Basically, we're just iterating over some maps here...
func calculateInfraCost(pricingData graphQLHTTPResponseBody, terraformResources resourceMap) ([]resourceCostMap, error) {
	var returnArray []resourceCostMap
	var oneDedicatedEc2 = false

	for resourceName, resourceTypes := range terraformResources.Resources {
		var resourceSpecificCostMap resourceCostMap
		resourceSpecificCostMap.Name = resourceName
		resourceSpecificCostMap.Resources = map[string]float32{"": 0.00}
		for resourceType, count := range resourceTypes {
			var alias string
			alias = strings.Replace(strings.Replace(strings.Replace(resourceType, ".", "_", -1), ",", "_", -1), "-", "_", -1)
			var price float64
			var err error
			for _, element := range pricingData.Data[alias] {
				price, err = strconv.ParseFloat(element.PricePerUnit, 32)
				if err != nil {
					return []resourceCostMap{}, err
				}
			}
			resourceSpecificCostMap.Resources[resourceType] = (float32(price) * float32(count))
			if oneDedicatedEc2 == false && resourceName == "aws_instance" && strings.Split(resourceType, ",")[1] == "Dedicated" {
				resourceSpecificCostMap.Resources["DedicatedPerRegionFee"] = 2.00
				oneDedicatedEc2 = true
			}
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

func processTerraformPlan(masterResourceMap resourceMap, planFile string) (resourceMap, error) {
	file, err := planfile.Open(planFile)
	if err != nil {
		return masterResourceMap, err
	}

	plan, err := file.ReadPlan()
	if err != nil {
		return masterResourceMap, err
	}

	decodedStruct, err := plan.Changes.Resources[1].Decode(cty.DynamicVal)

	fmt.Printf("%+v\n", decodedStruct)

	os.Exit(0)

	for _, resource := range plan.Changes.Resources {
		fmt.Printf("%+v\n", resource.Addr)
	}

	/*
		for resource, instanceChanges := range plan.Changes.Resources {
			resourceType := strings.Split(resource, ".")[0]
			switch resourceType {
			case "aws_instance":
				var resourceMapKey string
				if instanceChanges.Attributes["tenancy"].New == "dedicated" {
					resourceMapKey = instanceChanges.Attributes["instance_type"].New + ",Dedicated"
				} else {
					resourceMapKey = instanceChanges.Attributes["instance_type"].New + ",Shared"
				}
				masterResourceMap = countResource(masterResourceMap, resourceType, resourceMapKey)
			case "aws_db_instance":
				var resourceMapKey string
				if instanceChanges.Attributes["multi_az"].New == "true" {
					resourceMapKey = instanceChanges.Attributes["instance_class"].New + "," + instanceChanges.Attributes["engine"].New + ",Multi-AZ"
				} else {
					resourceMapKey = instanceChanges.Attributes["instance_class"].New + "," + instanceChanges.Attributes["engine"].New + ",Single-AZ"
				}
				masterResourceMap = countResource(masterResourceMap, resourceType, resourceMapKey)
			default:
				fmt.Println("resource type not recognized: ", resourceType)
			}
		}
	*/

	return masterResourceMap, nil
}

// This takes a terraform file and adds it to the global resource map used to shape the GraphQL query

func countResource(masterResourceMap resourceMap, resourceType string, resourceDescription string) resourceMap {
	if count := masterResourceMap.Resources[resourceType][resourceDescription]; count == 0 {
		masterResourceMap.Resources[resourceType][resourceDescription] = 1
	} else {
		masterResourceMap.Resources[resourceType][resourceDescription] = count + 1
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
					var alias string
					alias = strings.Replace(strings.Replace(strings.Replace(resourceType, ".", "_", -1), ",", "_", -1), "-", "_", -1)
					switch resource {
					case "aws_instance":
						ec2Instance := strings.Split(resourceType, ",")
						instanceType := ec2Instance[0]
						tenancy := ec2Instance[1]
						graphQLQueryString = graphQLQueryString + " " + fmt.Sprintf(queryStringTemplate, alias, region, "OnDemand", instanceType, tenancy)
					case "aws_db_instance":
						rdsInstance := strings.Split(resourceType, ",")
						instanceClass := rdsInstance[0]
						engine := rdsInstance[1]
						deploymentOption := rdsInstance[2]
						graphQLQueryString = graphQLQueryString + " " + fmt.Sprintf(queryStringTemplate, alias, region, "OnDemand", instanceClass, deploymentOption, engine)
					}
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
