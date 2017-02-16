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

	arrayOfResources := decodedOutput["resource"].([]map[string]interface{})

	for _, resource := range arrayOfResources {
		for key := range resource {
			fmt.Println("Key: ", key)

			switch key {
			case "aws_instance":
				fmt.Println("This is an AWS Instance")
				resourceKeys := resource[key].([]map[string]interface{})
				for resourceKey := range resourceKeys[0] {
					fmt.Println("Instance Size: ", resourceKeys[0][resourceKey].([]map[string]interface{})[0]["instance_type"])
				}
			default:
				fmt.Println("resource type not recognized")
			}
		}
	}
}
