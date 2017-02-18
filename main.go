package main

import (
	"fmt"
	"github.com/hashicorp/hcl"
	"io/ioutil"
)

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
	fmt.Printf("%+v\n", masterResourceMap)
}

// This doesn't need a pointer, or to return anything, because maps in go are always passed by reference
func countResource(resourceMap map[string]map[string]int, resourceName string, resourceType string) {
	if count := resourceMap[resourceName][resourceType]; count == 0 {
		resourceMap[resourceName][resourceType] = 1
	} else {
		resourceMap[resourceName][resourceType] = count + 1
	}
}

// func generateGraphQLQuery(masterResourceMap) string {
// }
