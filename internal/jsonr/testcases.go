package jsonr

type testCase struct {
	name    string
	jsonR   string
	want    Value
	wantErr bool
}

var testCases = []testCase{
	{
		name: "Valid JSON-R",
		jsonR: `
		{
			"name": { "value": "John Doe", "timestamp": 1678886400 },
			"age": { "value": 30, "timestamp": 1678886400 },
			"is-married": { "value": true, "timestamp": 1678886400 },
			"address": {
				"street": { "value": "123 Main St", "timestamp": 1678886400 },
				"city": { "value": "Anytown", "timestamp": 1678886400 }
			},
			"hobbies": [
				{ "value": "reading", "timestamp": 1678886400 },
				{ "value": "hiking", "timestamp": 1678886400 }
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
		jsonR:   `{ "name": "John Doe", }`,
		want:    nil,
		wantErr: true,
	},
	{
		name:    "Missing Value Key",
		jsonR:   `{"name": {"timestamp": 1678886400}}`,
		want:    nil,
		wantErr: true,
	},
	{
		name:    "Missing Timestamp Key",
		jsonR:   `{"name": {"value": "John Doe"}}`,
		want:    nil,
		wantErr: true,
	},
}
