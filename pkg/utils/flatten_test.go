package utils

import (
	"testing"
)

type TestStruct struct {
	Inner1Struct
	*Inner2Struct
	*Inner3Struct

	Field1 int
	Field2 string

	AStruct   *AStruct
	NilStruct *NilStruct

	MapB        map[string]MapBStruct
	MapCPointer map[int]*MapCStruct
	NilMap      map[int]*NilMapStruct

	SliceD   []SliceDStruct
	SliceE   []int
	NilSlice []int
}

type AStruct struct {
	AField    AAStruct
	NilStruct *NilStruct

	*Inner3Struct
}
type AAStruct struct{ AAField int }
type NilStruct struct{ NilField int }
type MapBStruct struct{ MapBField int }
type MapCStruct struct{ MapCField int }
type NilMapStruct struct{ NilMapField int }
type SliceDStruct struct{ SliceDField int }
type Inner1Struct struct{ Inner1Field int }
type Inner2Struct struct{ Inner2Field int }
type Inner3Struct struct{ Inner3Field int }

func TestFlatten(t *testing.T) {
	s := &TestStruct{
		Inner1Struct: Inner1Struct{
			Inner1Field: 1,
		},
		Inner2Struct: &Inner2Struct{
			Inner2Field: 2,
		},
		Inner3Struct: nil,
		Field1:       3,
		Field2:       "three",
		AStruct: &AStruct{
			AField:       AAStruct{AAField: 4},
			NilStruct:    nil,
			Inner3Struct: &Inner3Struct{Inner3Field: 3},
		},
		NilStruct: nil,
		MapB: map[string]MapBStruct{
			"b1": {
				MapBField: 5,
			},
			"b2": {
				MapBField: 6,
			},
		},
		MapCPointer: map[int]*MapCStruct{
			7: {
				MapCField: 7,
			},
			8: nil,
		},
		NilMap: nil,
		SliceD: []SliceDStruct{
			{
				SliceDField: 8,
			},
			{
				SliceDField: 9,
			},
		},
		SliceE:   []int{10, 11},
		NilSlice: nil,
	}

	flattened := Flatten(s, ".")

	if _, ok := flattened["Inner1Struct.Inner1Field"]; ok {
		t.Errorf("Inner1Struct.Inner1Field should be in flattened")
	} else if _, ok := flattened["Inner1Field"]; !ok {
		t.Errorf("Inner1Field should be in flattened")
	}

	if _, ok := flattened["Inner2Struct.Inner2Field"]; ok {
		t.Errorf("Inner2Struct.Inner2Field should be in flattened")
	} else if _, ok := flattened["Inner2Field"]; !ok {
		t.Errorf("Inner2Field should be in flattened")
	}

	if _, ok := flattened["Inner3Struct.Inner3Field"]; ok {
		t.Errorf("Inner3Struct should be not exist in flattened")
	}

	if _, ok := flattened["Field1"]; !ok {
		t.Errorf("Field1 should be in flattened")
	}

	if _, ok := flattened["AStruct.AField.AAField"]; !ok {
		t.Errorf("AStruct.AField.AAField should be in flattened")
	} else if _, ok := flattened["AStruct.NilStruct.NilField"]; ok {
		t.Errorf("AStruct.NilStruct.NilField should not exist in flattened")
	} else if _, ok := flattened["AStruct.Inner3Field"]; !ok {
		t.Errorf("AStruct.Inner3Field should be in flattened")
	}

	if _, ok := flattened["NilStruct.NilField"]; ok {
		t.Errorf("NilStruct.NilField should not exist in flattened")
	}

	if _, ok := flattened["MapB.b1.MapBField"]; !ok {
		t.Errorf("MapB.b1.MapBField should be in flattened")
	}

	if _, ok := flattened["MapCPointer.7.MapCField"]; !ok {
		t.Errorf("MapCPointer.7.MapCField should be in flattened")
	}

	if _, ok := flattened["MapCPointer.8.MapCField"]; ok {
		t.Errorf("MapCPointer.8.MapCField should not exist in flattened")
	}

	if _, ok := flattened["NilMap.0"]; ok {
		t.Errorf("NilMap.0 should not exist in flattened")
	}

	if _, ok := flattened["SliceD.0.SliceDField"]; !ok {
		t.Errorf("SliceD.0.SliceDField should be in flattened")
	}

	if _, ok := flattened["SliceE.0"]; ok {
		t.Errorf("SliceE.0 should not exist in flattened")
	}

}
