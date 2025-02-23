//
// testcases.go
//
// Test cases for testing tson package.
//
// Author: Karu (@karu-rress)
//

package tson

type testCase struct {
	name    string
	tson   string
	want    Value
	wantErr bool
}

var testCases = []testCase{
	{
		name: "Valid TSON",
		tson: `
		{
			"name" <1678886400>: "John Doe",
			"age" <1678886400>: 30,
			"is-married" <1678886400>: true,
			"address": {
				"street" <1678886400>: "123 Main St",
				"city" <1678886400>: "Anytown"
			},
			"hobbies": [
				<1678886400> "reading",
				<1678886400> "hiking"
			]
		}`,
		want: Object{
			"name":       Leaf[string]{Value: "John Doe", Timestamp: 1678886400},
			"age":        Leaf[float64]{Value: 30, Timestamp: 1678886400},
			"is-married": Leaf[bool]{Value: true, Timestamp: 1678886400},
			"address": Object{
				"street": Leaf[string]{Value: "123 Main St", Timestamp: 1678886400},
				"city":   Leaf[string]{Value: "Anytown", Timestamp: 1678886400},
			},
			"hobbies": Array{
				Leaf[string]{Value: "reading", Timestamp: 1678886400},
				Leaf[string]{Value: "hiking", Timestamp: 1678886400},
			},
		},
		wantErr: false,
	},
	{
		name:    "Invalid JSON",
		tson:   `{ "name": "John Doe", }`,
		want:    nil,
		wantErr: true,
	},
	{
		name:    "Missing Value",
		tson:   `{"name" <1678886400>: }`,
		want:    nil,
		wantErr: true,
	},
	{
		name:    "Missing Timestamp",
		tson:   `{"name": "John Doe"}`,
		want:    nil,
		wantErr: true,
	},
}
