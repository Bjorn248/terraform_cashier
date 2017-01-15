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
		}
	}
}
