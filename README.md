# cborquery

# Overview

This is a XPath query package for parsing and querying
[Concise Binary Object Representation (CBOR)](https://cbor.io) messages
using the [xpath](https://github.com/antchfx/xpath) package. The query package
is writing in pure go.

cborquery helps you easily and flexibly extract data from CBOR messages using
XPath queries without using pre-defined object structures.

### Install Package

```
go get github.com/srebhan/cborquery
```

## Get Started

Below is some example code to show how to use the package. Here we expect a
message similar to the following JSON data

```json
{
    "person":{
        "name":"John",
        "age":31,
        "female":false,
        "city":null,
        "hobbies":[
            "coding",
            "eating",
            "football"
        ]
    }
}
```

Using the xpath
syntax you are able to access specific fields of a CBOR message

```go
package main

import (
	"fmt"
	"strings"

	"github.com/srebhan/cborquery"
)

func main() {
    // Get a CBOR message from somewhere
    buf, err := os.ReadFile("mymessage.cbor")
	if err != nil {
		panic(err)
	}

    // Parse the message to be able to query the data
	doc, err := cborquery.Parse(bytes.NewBuffer(s))
	if err != nil {
		panic(err)
	}

	// xpath query
	age := cborquery.FindOne(doc, "age")
	// or
	age = cborquery.FindOne(doc, "person/age")
	fmt.Printf("%#v[%T]\n", age.Value(), age.Value())

	hobbies := cborquery.FindOne(doc, "//hobbies")
	fmt.Printf("%#v\n", hobbies.Value())
	firstHobby := cborquery.FindOne(doc, "//hobbies/*[1]")
	fmt.Printf("%#v\n", firstHobby.Value())
}
```

Iterating over the content

```go
package main

import (
	"fmt"
	"strings"

	"github.com/srebhan/cborquery"
)

func main() {
    // Get a CBOR message from somewhere
    buf, err := os.ReadFile("mymessage.cbor")
	if err != nil {
		panic(err)
	}

    doc, err := cborquery.Parse(strings.NewReader(s))
	if err != nil {
		panic(err)
	}
	// iterate all json objects from child ndoes.
	for _, n := range doc.ChildNodes() {
		fmt.Printf("%s: %v[%T]\n", n.Data, n.Value(), n.Value())
	}
}
```

For more information on XPath and supported features and function
see https://github.com/antchfx/xpath