package createschema

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/holgerjh/genjsonschema"
	"github.com/holgerjh/genjsonschema-cli/internal/merge"
	"gopkg.in/yaml.v2"
)

func TestCreateSchema(t *testing.T) {
	type given struct {
		config *genjsonschema.SchemaConfig
		inputs [][]byte //JSON or YAMl
	}

	tests := []struct {
		name          string
		given         given
		wantSchemaErr bool
		wantMergeErr  bool
	}{
		{
			name: "successfull generation 1",
			given: given{
				config: genjsonschema.NewDefaultSchemaConfig(),
				inputs: [][]byte{[]byte("42")},
			},
		},
		{
			name: "successfull generation 2",
			given: given{
				config: genjsonschema.NewDefaultSchemaConfig(),
				inputs: [][]byte{[]byte("{\"foo\": \"bar\"}")},
			},
		},
		{
			name: "non-default config",
			given: given{
				config: &genjsonschema.SchemaConfig{
					AdditionalProperties: true,
					RequireAllProperties: false,
				},
				inputs: [][]byte{[]byte("42")},
			},
		},
		{
			name: "merge error",
			given: given{
				config: genjsonschema.NewDefaultSchemaConfig(),
				inputs: [][]byte{[]byte("42"), []byte("{\"foo\": \"bar\"}")},
			},
			wantMergeErr: true,
		},
		{
			name: "schema from merged",
			given: given{
				config: genjsonschema.NewDefaultSchemaConfig(),
				inputs: [][]byte{[]byte("{\"foo\": \"bar\"}"), []byte("{\"foo\": \"baz\"}")},
			},
		},
		{
			name: "schema error",
			given: given{
				config: genjsonschema.NewDefaultSchemaConfig(),
				inputs: [][]byte{[]byte("42: \"bar\"")}, // int keys are not allowed
			},
			wantSchemaErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			merged, err := merge.MergeAllYAML(test.given.inputs...)
			if err != nil {
				if !test.wantMergeErr {
					t.Errorf("Got merge error but expected none: %v", err)
				}
				return
			}
			if test.wantMergeErr {
				t.Errorf("Got no merge error but expected one")
			}
			mergedAsYaml, err := yaml.Marshal(merged)
			if err != nil {
				t.Fatalf("%v", err)
			}
			want, err := genjsonschema.GenerateFromYAML(mergedAsYaml, test.given.config)
			if err != nil {
				if !test.wantSchemaErr {
					t.Fatalf("Got schema error but test is misconfigured and expected none: %v", err)
				} // do not return here, the "real" check is below
			}

			readers := make([]io.Reader, 0)
			for _, v := range test.given.inputs {
				readers = append(readers, bufio.NewReader(bytes.NewReader(v)))
			}

			got, err := CreateSchemaFromFiles(test.given.config, readers, false)
			if err != nil {
				if !test.wantSchemaErr {
					t.Errorf("Got schema error but expected none: %v", err)
				}
				return
			}
			if test.wantSchemaErr {
				t.Errorf("Got no schema error but expected one")
			}
			if delta := cmp.Diff(string(want), string(got)); delta != "" {
				t.Errorf("Wanted %s but got %s, delta %s", string(want), string(got), delta)
			}

		})
	}

}

func TestOnlyMerge(t *testing.T) {
	given := []io.Reader{
		bytes.NewReader([]byte(`{"foo": "bar"}`)),
		bytes.NewReader([]byte(`{"bar": "baz"}`)),
	}
	want := `{"foo": "bar", "bar": "baz"}`
	got, err := CreateSchemaFromFiles(genjsonschema.NewDefaultSchemaConfig(), given, true)
	if err != nil {
		t.Errorf("failed creating schema: %v", err)
	}
	var objWant, objGot interface{}
	if err := json.Unmarshal([]byte(want), &objWant); err != nil {
		t.Fatalf("failed generating json: %v", err)
	}
	if err := yaml.Unmarshal([]byte(got), &objGot); err != nil {
		t.Fatalf("failed generating json: %v", err)
	}

	rawWant, err := json.Marshal(objWant)
	if err != nil {
		t.Fatalf("%v", err)
	}
	rawGot, err := json.Marshal(objWant)
	if err != nil {
		t.Fatalf("%v", err)
	}

	if diff := cmp.Diff(rawWant, rawGot); diff != "" {
		t.Errorf("wanted %s but got %s, diff: %s", string(want), string(got), diff)
	}
}
