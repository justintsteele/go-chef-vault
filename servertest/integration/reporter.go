package integration

import (
	"encoding/json"
	"fmt"
)

func report(messge string, result any) {
	fmt.Printf("%s:\n", messge)
	if result == nil {
		fmt.Println("(no result)")
		return
	}

	jsonRes, err := json.Marshal(result)
	if err != nil {
		fmt.Printf("error marshaling result: %v\n", err)
		return
	}

	fmt.Println(string(jsonRes))
	fmt.Println()
}
