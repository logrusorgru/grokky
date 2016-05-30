package grokky

import (
	"fmt"
	"log"
)

func ExamplePattern_Parse() {
	h := New()
	h.Add("WORD", `\w+`)
	h.Add("NUMBER", `\d+`)
	p, err := h.Compile("%{WORD:name}/%{NUMBER:age}")
	if err != nil {
		log.Fatal(err)
	}
	result := p.Parse("Alice/15")
	fmt.Println("Name:", result["name"])
	fmt.Println("Age:", result["age"])
	// Output:
	// Name: Alice
	// Age: 15
}
