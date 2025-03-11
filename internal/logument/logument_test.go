package logument_test

import (
	"testing"

	"github.com/CAU-CPSS/logument/internal/logument"
	"github.com/CAU-CPSS/logument/internal/tson"
	"github.com/CAU-CPSS/logument/internal/tsonpatch"
	"github.com/davecgh/go-spew/spew"
)

const initSnapshot = `{
    "vehicleId" <1700000000>: "ABC1234",
    "speed" <1700000000>: 72.5,
    "engineOn" <1700000000>: true,
    "location": {
        "latitude" <1700000000>: 37.7749,
        "longitude" <1700000000>: -122.4194
    },
    "tirePressure": [
        <> 32.1,
        <> 31.8,
        <1700000000> 32.0,
        <1700000000> 31.9
    ]
}`

var Patches = []string{
	`[
		{ "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 1800000000 },
		{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 1800000000 },
		{ "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 1900000000 }
	]`,
	`[
		{ "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 }
	]`,
	`[
		{ "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 2100000000 },
		{ "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2100000000 }
	]`,
	`[
		{ "op": "replace", "path": "/tirePressure/2", "value": 33.7, "timestamp": 2300000000 },
		{ "op": "replace", "path": "/speed", "value": 94.9, "timestamp": 2400000000 }
	]`,
}

func TestCreate(t *testing.T) {
	t.Log("Make a new Logument document\n")

	// use only string format
	t.Log("Make a Logument with string format\n")
	lgm := logument.NewLogument(initSnapshot, Patches[0])
	t.Log(spew.Sdump(lgm))

	// use Snapshot and Patches format
	t.Log("Make a Logument with Snapshot format\n")
	var ss tson.Tson
	if err := tson.Unmarshal([]byte(initSnapshot), &ss); err != nil {
		panic(err)
	}
	pp, err := tsonpatch.Unmarshal([]byte(Patches[0]))
	if err != nil {
		t.Error(err)
	}
	lgmWithFormatdata := logument.NewLogument(ss, pp)
	t.Log(spew.Sdump(lgmWithFormatdata))

	// use []Patches format
	t.Log("Make a Logument with []Patches format\n")
	pp2, err := tsonpatch.Unmarshal([]byte(Patches[1]))
	if err != nil {
		t.Error(err)
	}
	lgmWithPatches := logument.NewLogument(ss, []tsonpatch.Patch{pp, pp2})
	t.Log(spew.Sdump(lgmWithPatches))
}

func TestStore(t *testing.T) {
	t.Log("Store patches to the pool\n")
	lgm := logument.NewLogument(initSnapshot, nil)

	lgm.Store(Patches[0])
	t.Log(spew.Sdump(lgm))
	lgm.Store(Patches[1])
	t.Log(spew.Sdump(lgm))
}

func TestApply(t *testing.T) {
	t.Log("Apply patches\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	t.Log(spew.Sdump(lgm))

	err := lgm.Append()
	if err != nil {
		t.Error(err)
	}
	t.Log(spew.Sdump(lgm))
}

func TestSnapshot(t *testing.T) {
	t.Log("Take a snapshot\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	lgm.Append()


	lgm.Print()

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

func TestTemporalSnapshot(t *testing.T) {
	t.Log("Take a timed snapshot\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	lgm.Append()


	// Take a snapshot already taken
	// snapshot := lgm.TimedSnapshot(1700000000)
	// t.Log(spew.Sdump(snapshot))

	// Take a snapshot
	snapshot := lgm.TemporalSnapshot(1900000000)
	t.Log(spew.Sdump(snapshot))

	// Requests exceeding latest version
	// snapshot = lgm.TimedSnapshot(2100000000)
	// t.Log(spew.Sdump(snapshot))
}

func TestSlice(t *testing.T) {
	t.Log("Slice Logument\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	err := lgm.Append() // version 1
	if err != nil {
		t.Error(err)
	} 
	lgm.Store(Patches[2])
	err = lgm.Append() // version 2
	if err != nil {
		t.Error(err)
	}
	lgm.Store(Patches[3])
	err = lgm.Append() // version 3
	if err != nil {
		t.Error(err)
	}

	// Slice patches
	lgmSubset := lgm.Slice(1, 3)
	lgmSubset.Print()
}

func TestTrack(t *testing.T) {
	t.Log("Track changes\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	lgm.Append() // version 1
	lgm.Store(Patches[2])
	lgm.Append() // version 2
	lgm.Store(Patches[3])
	lgm.Append() // version 3

	// Track changes
	// Expected:
	// { "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 2100000000 },
	// { "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2100000000 },
	// { "op": "replace", "path": "/tirePressure/2", "value": 33.7, "timestamp": 2300000000 },
	// { "op": "replace", "path": "/speed", "value": 94.9, "timestamp": 2400000000 }
	changes := lgm.Track(2, 3)
	t.Log(spew.Sdump(changes))
}

func TestTemporalTrack(t *testing.T) {
	t.Log("Track changes\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	lgm.Append() // version 1
	lgm.Store(Patches[2])
	lgm.Append() // version 2
	lgm.Store(Patches[3])
	lgm.Append() // version 3

	// Track changes
	// Expected:
	// { "op": "replace", "path": "/tirePressure/0", "value": 35.1, "timestamp": 1900000000 },
	// { "op": "replace", "path": "/engineOn", "value": false, "timestamp": 2000000000 },
	// { "op": "replace", "path": "/location/latitude", "value": 43.9409, "timestamp": 2100000000 },
	// { "op": "replace", "path": "/location/longitude", "value": -150.4194, "timestamp": 2100000000 }
	changes := lgm.TemporalTrack(1900000000, 2100000000)
	t.Log(spew.Sdump(changes))
}

func TestSet(t *testing.T) {
	t.Log("Set a value\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	lgm.Append()

	lgm.Store(Patches[2])
	lgm.Append()

	lgm.Set(2, "replace", "/location/latitude", 43.9409)
	lgm.Print()
}

func TestValidSet(t *testing.T) {
	t.Log("Set a value\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	lgm.Append()

	lgm.Store(Patches[2])
	lgm.Append()

	lgm.TestSet(2, "replace", "/location/latitude", 42.4242)
	lgm.Print()
}

func TestCompact(t *testing.T) {
	t.Log("Compact patches\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	lgm.Append()

	lgm.Store(Patches[2])
	lgm.Append()

	lgm.Compact("/location")
	lgm.Print()
}

func TestHistory(t *testing.T) {
	t.Log("Show history\n")
	lgm := logument.NewLogument(initSnapshot, nil)
	lgm.Store(Patches[0])
	lgm.Store(Patches[1])
	lgm.Append()

	lgm.Store(Patches[2])
	lgm.Append()

	// Show history of the "/location"
	his := lgm.History("/location")
	t.Log(spew.Sdump(his))
}
