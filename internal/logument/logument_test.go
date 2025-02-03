package logument_test

import (
	"testing"

	"github.com/CAU-CPSS/logument/internal/jsonpatch"
	"github.com/CAU-CPSS/logument/internal/jsonr"
	"github.com/CAU-CPSS/logument/internal/logument"
	"github.com/davecgh/go-spew/spew"
)

const initSnapshot = `{
    "vehicleId": { "value": "ABC1234", "timestamp": 1700000000 },
    "speed": { "value": 72.5, "timestamp": 1700000000 },
    "engineOn": { "value": true, "timestamp": 1700000000 },
    "location": { "latitude": { "value": 37.7749, "timestamp": 1700000000 },
        "longitude": { "value": -122.4194, "timestamp": 1700000000 } },
    "tirePressure": [
        { "value": 32.1, "timestamp": 1700000000 },
        { "value": 31.8, "timestamp": 1700000000 },
        { "value": 32.0, "timestamp": 1700000000 },
        { "value": 31.9, "timestamp": 1700000000 }
    ] }`

const initPatch = `[
	{ "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 1800000000 },	
	{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 1800000000 },
	{ "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 1900000000 }
]`

const secondPatch = `[
	{ "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 }
]`

const thirdPatch = `[
	{ "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 2100000000 },	
	{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2100000000 }
]`

func TestCreate(t *testing.T) {
	t.Log("Make a new Logument document\n")

	// use only string format
	t.Log("Make a Logument with string format\n")
	lgm := logument.NewLogument(initSnapshot, initPatch)
	t.Log(spew.Sdump(lgm))

	// use Snapshot and Patches format
	t.Log("Make a Logument with Snapshot format\n")
	var ss jsonr.JsonR
	err := jsonr.Unmarshal([]byte(initSnapshot), &ss)
	if err != nil {
		panic(err)
	}
	pp, err := jsonpatch.Unmarshal([]byte(initPatch))
	if err != nil {
		t.Error(err)
	}
	lgmWithFormatdata := logument.NewLogument(ss, pp)
	t.Log(spew.Sdump(lgmWithFormatdata))

	// use []Patches format
	t.Log("Make a Logument with []Patches format\n")
	pp2, err := jsonpatch.Unmarshal([]byte(secondPatch))
	if err != nil {
		t.Error(err)
	}
	lgmWithPatches := logument.NewLogument(ss, []jsonpatch.Patch{pp, pp2})
	t.Log(spew.Sdump(lgmWithPatches))
}

func TestStore(t *testing.T) {
	t.Log("Store patches to the pool\n")
	lgm := logument.NewLogument(initSnapshot, nil)

	lgm.Store(initPatch)
	t.Log(spew.Sdump(lgm))
	lgm.Store(secondPatch)
	t.Log(spew.Sdump(lgm))
}

func TestApply(t *testing.T) {
	t.Log("Apply patches\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(initPatch)
	lgm.Store(secondPatch)
	t.Log(spew.Sdump(lgm))

	lgm.Apply()
	t.Log(spew.Sdump(lgm))
}

func TestSnapshot(t *testing.T) {
	t.Log("Take a snapshot\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(initPatch)
	lgm.Store(secondPatch)
	lgm.Apply()

	// Take a snapshot already taken
	snapshot := lgm.Snapshot(0)
	t.Log(spew.Sdump(snapshot))

	// Take a snapshot
	snapshot = lgm.Snapshot(1)
	t.Log(spew.Sdump(snapshot))

	// Requests exceeding latest version
	// snapshot = lgm.Snapshot(3)
	// t.Log(spew.Sdump(snapshot))
}

func TestTimedSnapshot(t *testing.T) {
	t.Log("Take a timed snapshot\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(initPatch)
	lgm.Store(secondPatch)
	lgm.Apply()

	// Take a snapshot already taken
	// snapshot := lgm.TimedSnapshot(1700000000)
	// t.Log(spew.Sdump(snapshot))

	// Take a snapshot
	snapshot := lgm.TimedSnapshot(1900000000)
	t.Log(spew.Sdump(snapshot))

	// // Requests exceeding latest version
	// snapshot = lgm.TimedSnapshot(2100000000)
	// t.Log(spew.Sdump(snapshot))
}

func TestCompact(t *testing.T) {
	t.Log("Compact patches\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(initPatch)
	lgm.Store(secondPatch)
	lgm.Apply()

	lgm.Store(thirdPatch)
	lgm.Apply()

	lgm.Compact("/location")
	lgm.Print()
}

func TestHistory(t *testing.T) {
	t.Log("Show history\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(initPatch)
	lgm.Store(secondPatch)
	lgm.Apply()

	lgm.Store(thirdPatch)
	lgm.Apply()

	// Show history of the "/location"
	his := lgm.History("/location")
	t.Log(spew.Sdump(his))
}