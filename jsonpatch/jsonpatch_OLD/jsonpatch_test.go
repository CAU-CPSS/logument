package jsonpatch_test

import (
	"encoding/json"
	"jsonpatch-plus"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Test codes
var logumentA = `{"a": {"Value": 100, "Timestamp": 1111111111}, "b": {"Value": 200, "Timestamp": 1111111111}, "c": {"Value": "hello", "Timestamp": 1111111111}}`
var logumentB = `{"a": {"Value": 100, "Timestamp": 2222222222}, "b": {"Value": 200, "Timestamp": 2222222222}, "c": {"Value": "goodbye", "Timestamp": 2222222222}}`
var logumentC = `{"a": {"Value": 100, "Timestamp": 3333333333}, "b": {"Value": 100, "Timestamp": 3333333333}, "c": {"Value": "hello", "Timestamp": 3333333333}}`
var logumentD = `{"a": {"Value": 100, "Timestamp": 4444444444}, "b": {"Value": 200, "Timestamp": 4444444444}, "c": {"Value": "hello", "Timestamp": 4444444444}, "d": {"Value": "foo", "Timestamp": 4444444444}}`
var logumentE = `{"a": {"Value": 100, "Timestamp": 5555555555}, "b": {"Value": 200, "Timestamp": 5555555555}}`
var logumentF = `{"a": {"Value": 100, "Timestamp": 6666666666}, "b": {"Value": 100, "Timestamp": 6666666666}, "d": {"Value": "foo", "Timestamp": 6666666666}}`
var logumentG = `{"a": {"Value": 100, "Timestamp": 7777777777}, "b": null, "d": {"Value": "foo", "Timestamp": 7777777777}}`
var emptyLogument = `{}`

var arraySrc = `{
  "spec": {
    "loadBalancerSourceRanges": {
      "Value": [
        "192.101.0.0/16",
        "192.0.0.0/24"
      ],
	  "Timestamp": 1111111111
    }
  }
}
`

var arrayDst = `{
  "spec": {
    "loadBalancerSourceRanges": {
	  "Value": [
	    "192.101.0.0/16",
	  ],
	  "Timestamp": 2222222222
    }
  }
}
`

var complexLogumentBase = `{
  "a": {
    "Value": 100, 
	"Timestamp": 1111111111
  }, 
  "b": {
    "Value": [
	  {
	    "c1": {
	      "Value": "hello", 
		  "Timestamp": 1111111111
	    }, 
	    "d1": {
	      "Value": "foo", 
		  "Timestamp": 1111111111
	    }
	  }, {
	    "c2": {
	      "Value": "hello2", 
	      "Timestamp": 1111111111
	    }, 
	    "d2": {
	      "Value": "foo2",
		  "Timestamp": 1111111111
	    }
	  }
	],
	"Timestamp": 1111111111
  },
  "e": {
    "Value": {
	  "f": {
	    "Value": 200,
		"Timestamp": 1111111111
	  }, 
	  "g": {
	    "Value": "h",
		"Timestamp": 1111111111
	  },
	  "i": {
	    "Value": "j",
		"Timestamp": 1111111111
	  }
	},
	"Timestamp": 1111111111
  }
}`

// complexLogument 다시 만들어야 함
var complexLogumentA = `{"a": {"Value": 100, "Timestamp": 2222222222}, "b": {"Value": [{"c1": {"Value": "goodbye", "Timestamp": 2222222222}, "d1": {"Value": "foo", "Timestamp": 2222222222}},{"c2": {"Value": "hello2", "Timestamp": 2222222222}, "d2":{"Value": "foo2", "Timestamp": 2222222222}} ], "Timestamp": 2222222222}, "e": {"Value": {"f": {"Value": 200, "Timestamp": 2222222222}, "g": {"Value": "h", "Timestamp": 2222222222}, "i": {"Value": "j", "Timestamp": 2222222222}}, "Timestamp": 2222222222}}`
var complexLogumentB = `{"a": {"Value": 100, "Timestamp": 3333333333}, "b": {"Value": [{"c1": {"Value": "hello", "Timestamp": 3333333333}, "d1": {"Value": "foo", "Timestamp": 3333333333}},{"c2": {"Value": "hello2", "Timestamp": 3333333333}, "d2":{"Value": "foo2", "Timestamp": 3333333333}} ], "Timestamp": 3333333333}, "e": {"Value": {"f": {"Value": 100, "Timestamp": 3333333333}, "g": {"Value": "h", "Timestamp": 3333333333}, "i": {"Value": "j", "Timestamp": 3333333333}}, "Timestamp": 3333333333}}`
var complexLogumentC = `{"a": {"Value": 100, "Timestamp": 4444444444}, "b": {"Value": [{"c1": {"Value": "hello", "Timestamp": 4444444444}, "d1": {"Value": "foo", "Timestamp": 4444444444}},{"c2": {"Value": "hello2", "Timestamp": 4444444444}, "d2":{"Value": "foo2", "Timestamp": 4444444444}, "c3": {"Value": "hello3", "Timestamp": 4444444444}, "d3": {"Value": "foo3", "Timestamp": 4444444444}} ], "Timestamp": 4444444444}, "e": {"Value": {"f": {"Value": 200, "Timestamp": 4444444444}, "g": {"Value": "h", "Timestamp": 4444444444}, "i": {"Value": "j", "Timestamp": 4444444444}}, "Timestamp": 4444444444}, "k": {"Value": [{"l": {"Value": "m", "Timestamp": 4444444444}}, {"l": {"Value": "o", "Timestamp": 4444444444}}], "Timestamp": 4444444444}}`
var complexLogumentD = `{"a": {"Value": 100, "Timestamp": 5555555555}, "b": {"Value": [{"c1": {"Value": "hello", "Timestamp": 5555555555}, "d1": {"Value": "foo", "Timestamp": 5555555555}},{"c2": {"Value": "hello2", "Timestamp": 5555555555}, "d2":{"Value": "foo2", "Timestamp": 5555555555}} ], "Timestamp": 5555555555}, "e": {"Value": {"f": {"Value": 200, "Timestamp": 5555555555}, "g": {"Value": "h", "Timestamp": 5555555555}, "i": {"Value": "j", "Timestamp": 5555555555}}, "Timestamp": 5555555555}}`
var complexLogumentE = `{"a": {"Value": 100, "Timestamp": 6666666666}, "b": {"Value": [{"c1": {"Value": "hello", "Timestamp": 6666666666}, "d1": {"Value": "foo", "Timestamp": 6666666666}},{"c2": {"Value": "hello2", "Timestamp": 6666666666}, "d2":{"Value": "foo2", "Timestamp": 6666666666}} ], "Timestamp": 6666666666}, "e": {"Value": {"f": {"Value": 100, "Timestamp": 6666666666}, "g": {"Value": "h", "Timestamp": 6666666666}, "i": {"Value": "j", "Timestamp": 6666666666}}, "Timestamp": 6666666666}}`

var pointLogument = `{"type": {"Value": "Point", "Timestamp": 1111111111}, "coordinates": {"Value": [0.0, 1.0], "Timestamp": 1111111111}}`
var lineStringLogument = `{"type": {"Value": "LineString", "Timestamp": 1111111111}, "coordinates": {"Value": [[0.0, 1.0], [2.0, 3.0]], "Timestamp": 1111111111}}`

var hyperComplexBase = `
{
	"goods": [
	{
		"id": "0001",
		"type": "donut",
		"name": "Cake",
		"ppu": 0.55,
		"batters":
			{
				"batter":
					[
						{ "id": "1001", "type": "Regular" },
						{ "id": "1002", "type": "Chocolate" },
						{ "id": "1003", "type": "Blueberry" },
						{ "id": "1004", "type": "Devil's Food" }
					]
			},
		"topping":
			[
				{ "id": "5001", "type": "None" },
				{ "id": "5002", "type": "Glazed" },
				{ "id": "5005", "type": "Sugar" },
				{ "id": "5007", "type": "Powdered Sugar" },
				{ "id": "5006", "type": "Chocolate with Sprinkles" },
				{ "id": "5003", "type": "Chocolate" },
				{ "id": "5004", "type": "Maple" }
			]
	},
	{
		"id": "0002",
		"type": "donut",
		"name": "Raised",
		"ppu": 0.55,
		"batters":
			{
				"batter":
					[
						{ "id": "1001", "type": "Regular" }
					]
			},
		"topping":
			[
				{ "id": "5001", "type": "None" },
				{ "id": "5002", "type": "Glazed" },
				{ "id": "5005", "type": "Sugar" },
				{ "id": "5003", "type": "Chocolate" },
				{ "id": "5004", "type": "Maple" }
			]
	},
	{
		"id": "0003",
		"type": "donut",
		"name": "Old Fashioned",
		"ppu": 0.55,
		"batters":
			{
				"batter":
					[
						{ "id": "1001", "type": "Regular" },
						{ "id": "1002", "type": "Chocolate" }
					]
			},
		"topping":
			[
				{ "id": "5001", "type": "None" },
				{ "id": "5002", "type": "Glazed" },
				{ "id": "5003", "type": "Chocolate" },
				{ "id": "5004", "type": "Maple" }
			]
	}
]
}`

var hyperComplexA = `
{
	"goods": [
	{
		"id": "0001",
		"type": "donut",
		"name": "Cake",
		"ppu": 0.55,
		"batters":
			{
				"batter":
					[
						{ "id": "1001", "type": "Regular" },
						{ "id": "1002", "type": "Chocolate" },
						{ "id": "1003", "type": "Strawberry" },
						{ "id": "1004", "type": "Devil's Food" }
					]
			},
		"topping":
			[
				{ "id": "5001", "type": "None" },
				{ "id": "5002", "type": "Glazed" },
				{ "id": "5005", "type": "Sugar" },
				{ "id": "5007", "type": "Powdered Sugar" },
				{ "id": "5006", "type": "Chocolate with Sprinkles" },
				{ "id": "5003", "type": "Chocolate" },
				{ "id": "5004", "type": "Maple" }
			]
	},
	{
		"id": "0002",
		"type": "donut",
		"name": "Raised",
		"ppu": 0.55,
		"batters":
			{
				"batter":
					[
						{ "id": "1001", "type": "Regular" }
					]
			},
		"topping":
			[
				{ "id": "5001", "type": "None" },
				{ "id": "5002", "type": "Glazed" },
				{ "id": "5005", "type": "Sugar" },
				{ "id": "5003", "type": "Chocolate" },
				{ "id": "5004", "type": "Maple" }
			]
	},
	{
		"id": "0003",
		"type": "donut",
		"name": "Old Fashioned",
		"ppu": 0.55,
		"batters":
			{
				"batter":
					[
						{ "id": "1001", "type": "Regular" },
						{ "id": "1002", "type": "Chocolate" },
						{ "id": "1003", "type": "Vanilla" }
					]
			},
		"topping":
			[
				{ "id": "5001", "type": "None" },
				{ "id": "5002", "type": "Glazed" },
				{ "id": "5004", "type": "Maple" }
			]
	}
]
}`

var superComplexBase = `
{
	"annotations": {
		"annotation": [
			{
				"name": "version",
				"value": "8"
			},
			{
				"name": "versionTag",
				"value": "Published on May 13, 2015 at 8:48pm (MST)"
			}
		]
	},
	"attributes": {
		"attribute-key": [
			{
				"id": "3b05c943-d81a-436f-b242-8b519e7a6f30",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "d794c7ee-2a4b-4da4-bba7-e8b973d50c4b",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "a0259458-517c-480f-9f04-9b54b1b2af1f",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "9415f39d-c396-4458-9019-fc076c847964",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "0a2e49a9-8989-42fb-97da-cc66334f828b",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "27f5f14a-ea97-4feb-b22a-6ff754a31212",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "6f810508-4615-4fd0-9e87-80f9c94f9ad8",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "3451b1b2-7365-455c-8bb1-0b464d4d3ba1",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "a82ec957-8c26-41ea-8af6-6dd75c384801",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "736c5496-9a6e-4a82-aa00-456725796432",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "2d428b3c-9d3b-4ec1-bf98-e00673599d60",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "68566ebb-811d-4337-aba9-a8a8baf90e4b",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "ca88bab1-a1ea-40cc-8f96-96d1e9f1217d",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "c63a12c8-542d-47f3-bee1-30b5fe2b0690",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "cbd9e3bc-6a49-432a-a906-b1674c1de24c",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "03262f07-8a15-416d-a3f5-e2bf561c78f9",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "e5c93b87-83fc-45b6-b4d5-bf1e3f523075",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "72260ac5-3d51-49d7-bb31-f794dd129f1c",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "d856bde1-1b42-4935-9bee-c37e886c9ecf",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "62380509-bedf-4134-95c3-77ff377a4a6a",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "f4ed5ac9-b386-49a6-a0a0-6f3341ce9021",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "528d2bd2-87fe-4a49-954a-c93a03256929",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "ff8951f1-61a7-416b-9223-fac4bb6dac50",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "95c2b011-d782-4042-8a07-6aa4a5765c2e",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "dbe5837b-0624-4a05-91f3-67b5bd9b812a",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "13f198ed-82ab-4e51-8144-bfaa5bf77fd5",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "025312eb-12b6-47e6-9750-0fb31ddc2111",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "24292d58-db66-4ef3-8f4f-005d7b719433",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "22e5b5c4-821c-413a-a5b1-ab866d9a03bb",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "2fde0aac-df89-403d-998e-854b949c7b57",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "8b576876-5c16-4178-805e-24984c24fac3",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "415b7d2a-b362-4f1e-b83a-927802328ecb",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "8ef24fc2-ab25-4f22-9d9f-61902b49dc01",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "2299b09e-9f8e-4b79-a55c-a7edacde2c85",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "bf506538-f438-425c-be85-5aa2f9b075b8",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "2b501dc6-799d-4675-9144-fac77c50c57c",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "c0446da1-e069-417e-bd5a-34edcd028edc",
				"properties": {
					"visible": true
				}
			}
		]
	}
}`

var superComplexA = `
{
	"annotations": {
		"annotation": [
			{
				"name": "version",
				"value": "8"
			},
			{
				"name": "versionTag",
				"value": "Published on May 13, 2015 at 8:48pm (MST)"
			}
		]
	},
	"attributes": {
		"attribute-key": [
			{
				"id": "3b05c943-d81a-436f-b242-8b519e7a6f30",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "d794c7ee-2a4b-4da4-bba7-e8b973d50c4b",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "a0259458-517c-480f-9f04-9b54b1b2af1f",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "9415f39d-c396-4458-9019-fc076c847964",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "0a2e49a9-8989-42fb-97da-cc66334f828b",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "27f5f14a-ea97-4feb-b22a-6ff754a31212",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "6f810508-4615-4fd0-9e87-80f9c94f9ad8",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "3451b1b2-7365-455c-8bb1-0b464d4d3ba1",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "a82ec957-8c26-41ea-8af6-6dd75c384801",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "736c5496-9a6e-4a82-aa00-456725796432",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "2d428b3c-9d3b-4ec1-bf98-e00673599d60",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "68566ebb-811d-4337-aba9-a8a8baf90e4b",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "ca88bab1-a1ea-40cc-8f96-96d1e9f1217d",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "c63a12c8-542d-47f3-bee1-30b5fe2b0690",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "cbd9e3bc-6a49-432a-a906-b1674c1de24c",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "03262f07-8a15-416d-a3f5-e2bf561c78f9",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "e5c93b87-83fc-45b6-b4d5-bf1e3f523075",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "72260ac5-3d51-49d7-bb31-f794dd129f1c",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "d856bde1-1b42-4935-9bee-c37e886c9ecf",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "62380509-bedf-4134-95c3-77ff377a4a6a",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "f4ed5ac9-b386-49a6-a0a0-6f3341ce9021",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "528d2bd2-87fe-4a49-954a-c93a03256929",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "ff8951f1-61a7-416b-9223-fac4bb6dac50",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "95c2b011-d782-4042-8a07-6aa4a5765c2e",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "dbe5837b-0624-4a05-91f3-67b5bd9b812a",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "13f198ed-82ab-4e51-8144-bfaa5bf77fd5",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "025312eb-12b6-47e6-9750-0fb31ddc2111",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "24292d58-db66-4ef3-8f4f-005d7b719433",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "22e5b5c4-821c-413a-a5b1-ab866d9a03bb",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "2fde0aac-df89-403d-998e-854b949c7b57",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "8b576876-5c16-4178-805e-24984c24fac3",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "415b7d2a-b362-4f1e-b83a-927802328ecb",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "8ef24fc2-ab25-4f22-9d9f-61902b49dc01",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "2299b09e-9f8e-4b79-a55c-a7edacde2c85",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "bf506538-f438-425c-be85-5aa2f9b075b8",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "2b501dc6-799d-4675-9144-fac77c50c57c",
				"properties": {
					"visible": true
				}
			},
			{
				"id": "c0446da1-e069-417e-bd5a-34edcd028edc",
				"properties": {
					"visible": false
				}
			}
		]
	}
}`

var (
	oldDeployment = `{
  "apiVersion": "apps/v1beta1",
  "kind": "Deployment",
  "metadata": {
    "annotations": {
      "k8s.io/app": "busy-dep"
    }
  }
}`

	newDeployment = `{
  "apiVersion": "apps/v1beta1",
  "kind": "Deployment",
  "metadata": {
    "annotations": {
      "k8s.io/app": "busy-dep",
      "docker.com/commit": "github.com/myrepo#xyz"
    }
  }
}`
)

var (
	oldNestedObj = `{
  "apiVersion": "kubedb.com/v1alpha1",
  "kind": "Elasticsearch",
  "metadata": {
    "name": "quick-elasticsearch",
    "namespace": "demo"
  },
  "spec": {
    "doNotPause": true,
    "version": "5.6"
  }
}`

	newNestedObj = `{
  "apiVersion": "kubedb.com/v1alpha1",
  "kind": "Elasticsearch",
  "metadata": {
    "name": "quick-elasticsearch",
    "namespace": "demo"
  },
  "spec": {
    "doNotPause": true,
    "version": "5.6",
    "storageType": "Durable",
    "updateStrategy": {
      "type": "RollingUpdate"
    },
    "terminationPolicy": "Pause"
  }
}`
)

var (
	oldArray = `{
  "apiVersion": "kubedb.com/v1alpha1",
  "kind": "Elasticsearch",
  "metadata": {
    "name": "quick-elasticsearch",
    "namespace": "demo"
  },
  "spec": {
    "tolerations": [
      {
          "key": "node.kubernetes.io/key1",
          "operator": "Equal",
          "value": "value1",
          "effect": "NoSchedule"
      },
      {
          "key": "node.kubernetes.io/key2",
          "operator": "Equal",
          "value": "value2",
          "effect": "NoSchedule"
      },
      {
          "key": "node.kubernetes.io/not-ready",
          "operator": "Exists",
          "effect": "NoExecute",
          "tolerationSeconds": 300
      },
      {
          "key": "node.kubernetes.io/unreachable",
          "operator": "Exists",
          "effect": "NoExecute",
          "tolerationSeconds": 300
      }
    ]
  }
}`

	newArray = `{
  "apiVersion": "kubedb.com/v1alpha1",
  "kind": "Elasticsearch",
  "metadata": {
    "name": "quick-elasticsearch",
    "namespace": "demo"
  },
  "spec": {
    "tolerations": [
      {
          "key": "node.kubernetes.io/key2",
          "operator": "Equal",
          "value": "value2",
          "effect": "NoSchedule"
      },
      {
          "key": "node.kubernetes.io/key1",
          "operator": "Equal",
          "value": "value1",
          "effect": "NoSchedule"
      }
    ]
  }
}`
)

var (
	nullKeyA = `{
  "apiVersion": "cert-manager.io/v1",
  "kind": "CertificateRequest",
  "metadata": {
    "creationTimestamp": null,
    "name": "test-cr",
    "namespace": "default-unit-test-ns"
  },
  "spec": {
    "issuerRef": {
      "name": ""
    },
    "request": null
  },
  "status": {}
}`
	nullKeyB = `{
  "apiVersion": "cert-manager.io/v1",
  "kind": "CertificateRequest",
  "metadata": {
    "creationTimestamp": null,
    "name": "test-cr",
    "namespace": "default-unit-test-ns"
  },
  "spec": {
    "issuerRef": {
      "name": ""
    },
    "request": "bXV0YXRpb24gY2FsbGVk"
  },
  "status": {}
}`
)

var (
	emptyKeyA = `{"":[0]}`
	emptyKeyB = `{"":[]}`
)

var (
	specialChars = string([]byte{123, 34, 92, 98, 34, 58, 91, 93, 125})
)

func TestCreateSimplePatch(t *testing.T) {
	cases := []struct {
		name string
		src  string
		dst  string
	}{
		{"Simple:OneNullReplace", logumentF, logumentG},
		{"Simple:Same", logumentA, logumentA},
		{"Simple:OneStringReplace", logumentA, logumentB},
		{"Simple:OneIntReplace", logumentA, logumentC},
		{"Simple:OneAdd", logumentA, logumentD},
		{"Simple:OneRemove", logumentA, logumentE},
		{"Simple:VsEmpty", logumentA, emptyLogument},
	}

	for _, c := range cases {
		t.Run(c.name+"[src->dst]", func(t *testing.T) {
			check(t, c.src, c.dst)
		})
		t.Run(c.name+"[dst->src]", func(t *testing.T) {
			check(t, c.dst, c.src)
		})
	}
}

func TestCreatePatch(t *testing.T) {
	cases := []struct {
		name string
		src  string
		dst  string
	}{
		// simple
		{"Simple:OneNullReplace", logumentF, logumentG},
		{"Simple:Same", logumentA, logumentA},
		{"Simple:OneStringReplace", logumentA, logumentB},
		{"Simple:OneIntReplace", logumentA, logumentC},
		{"Simple:OneAdd", logumentA, logumentD},
		{"Simple:OneRemove", logumentA, logumentE},
		{"Simple:VsEmpty", logumentA, emptyLogument},
		// array types
		{"Array:Same", arraySrc, arraySrc},
		{"Array:BoolReplace", arraySrc, arrayDst},
		{"Array:AlmostSame", `{"Lines":[1,2,3,4,5,6,7,8,9,10]}`, `{"Lines":[2,3,4,5,6,7,8,9,10,11]}`},
		{"Array:Remove", `{"x":["A", "B", "C"]}`, `{"x":["D"]}`},
		{"Array:EditDistance", `{"letters":["A","B","C","D","E","F","G","H","I","J","K"]}`, `{"letters":["L","M","N"]}`},
		// complex types
		{"Complex:Same", complexBase, complexBase},
		{"Complex:OneStringReplaceInArray", complexBase, complexA},
		{"Complex:OneIntReplace", complexBase, complexB},
		{"Complex:OneAdd", complexBase, complexC},
		{"Complex:OneAddToArray", complexBase, complexC},
		{"Complex:VsEmpty", complexBase, emptyLogument},
		// geojson
		{"GeoJson:PointLineStringReplace", point, lineString},
		{"GeoJson:LineStringPointReplace", lineString, point},
		// HyperComplex
		{"HyperComplex:Same", hyperComplexBase, hyperComplexBase},
		{"HyperComplex:BoolReplace", hyperComplexBase, hyperComplexA},
		// SuperComplex
		{"SuperComplex:Same", superComplexBase, superComplexBase},
		{"SuperComplex:BoolReplace", superComplexBase, superComplexA},
		// map
		{"Kubernetes:Annotations", oldDeployment, newDeployment},
		// crd with nested object
		{"Nested Member Object", oldNestedObj, newNestedObj},
		// array with different order
		{"Different Array", oldArray, newArray},
		{"Array at root", `[{"asdf":"qwerty"}]`, `[{"asdf":"bla"},{"asdf":"zzz"}]`},
		{"Empty array at root", `[]`, `[{"asdf":"bla"},{"asdf":"zzz"}]`},
		{"Null Key uses replace operation", nullKeyA, nullKeyB},
		// empty key
		{"Empty key", emptyKeyA, emptyKeyB},
		// special chars
		{"Special chars", emptyLogument, specialChars},
	}

	for _, c := range cases {
		t.Run(c.name+"[src->dst]", func(t *testing.T) {
			check(t, c.src, c.dst)
		})
		t.Run(c.name+"[dst->src]", func(t *testing.T) {
			check(t, c.dst, c.src)
		})
	}
}

func check(t *testing.T, src, dst string) {
	patch, err := jsonpatch.CreatePatch([]byte(src), []byte(dst))
	assert.Nil(t, err)

	data, err := json.Marshal(patch)
	assert.Nil(t, err)

	p2, err := jp.DecodePatch(data)
	assert.Nil(t, err)

	d2, err := p2.Apply([]byte(src))
	assert.Nil(t, err)

	assert.JSONEq(t, dst, string(d2))
}
