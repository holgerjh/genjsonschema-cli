package merge

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/yaml.v2"
)

func TestMergeAll(t *testing.T) {
	tests := []struct {
		name    string
		given   []string // yaml inputs
		want    string   // expected output
		wantErr bool
	}{
		{
			name:  "single scalar",
			given: []string{"42"},
			want:  "42",
		},
		{
			name:  "merging two scalars (destructive)",
			given: []string{"42", "44"},
			want:  "44",
		},
		{
			name:  "number takes precendence over integer",
			given: []string{"42.4", "44"},
			want:  "42.4",
		},
		{
			name:  "merging two scalars (destructive)",
			given: []string{"\"foo\"", "\"bar\""},
			want:  "bar",
		},
		{
			name:    "reject merge integer and boolean",
			given:   []string{"42", "44", "false", "48"},
			wantErr: true,
		},
		{
			name:    "reject merge scalar with object",
			given:   []string{"42", `{"foo": "bar"}`},
			wantErr: true,
		},
		{
			name:    "reject merge list with object",
			given:   []string{"[42]", `{"foo": "bar"}`},
			wantErr: true,
		},
		{
			name:  "merge lists",
			given: []string{`["foo"]`, `["bar"]`, `["baz"]`, `["baz"]`},
			want:  `["baz", "bar", "foo"]`, // list order is unfortuantely checked by reflect.DeepEqual
		},
		{
			name: "merge objects",
			given: []string{
				"" +
					"foo: \"foovalue\"\n" +
					"bar: \"barvalue\"\n",
				"" +
					"bar: \"newbar\"\n" +
					"baz: \"newbaz\"\n",
			},
			want: "" +
				"foo: \"foovalue\"\n" +
				"bar: \"newbar\"\n" +
				"baz: \"newbaz\"\n",
		},
		{
			name: "complex nested merge with scalar, lists and objects",
			given: []string{
				"" +
					"outer:\n" +
					"  foo: \"bar\"\n" +
					"  numbers: [1, 2, 3]\n" +
					"  inner: {\"foo\": 42}\n",
				"" +
					"outer:\n" +
					"  foo: \"bar-replaced\"\n" +
					"  numbers: [3, 4, 5]\n" +
					"  inner: {\"bar\": \"baz\"}\n",
			},
			want: "" +
				"outer:\n" +
				"  foo: \"bar-replaced\"\n" +
				"  numbers: [3, 4, 5, 1, 2]\n" +
				"  inner: {\"foo\": 42, \"bar\": \"baz\"}\n",
		},
	}
	for _, tt := range tests {
		fmt.Println(tt.name)
		t.Run(tt.name, func(t *testing.T) {

			given := make([][]byte, len(tt.given))
			for i, v := range tt.given {
				given[i] = []byte(v)
			}

			got, err := MergeAllYAML(given...)

			if err == nil {
				if tt.wantErr { //got no error but wanted one
					t.Errorf("given %v wanted error but got result %v instead", tt.given, got)
					return
				}
			} else {
				if !tt.wantErr { //got an error but wanted none
					t.Errorf("given %v got error but expected none: %v", tt.given, err)
					return
				}
				return // got an error and wanted one, all ok
			}

			// we might end up with different datatypes for whole numbers (e.g. float64 and int)
			// so we translate everything back into YAML to normalize the representation
			gotYAML, err := yaml.Marshal(got)
			if err != nil {
				t.Errorf("%s", err)
				return
			}

			// translate "want" into interface and back into YAML
			var want interface{} = nil
			if tt.want != "" {
				err := yaml.Unmarshal([]byte(tt.want), &want)
				if err != nil {
					t.Fatalf("Unmarshalling 'want' failed: %v", err)
				}
			}
			wantYAML, err := yaml.Marshal(want)
			if err != nil {
				t.Fatalf("%s", err)
			}

			if diff := cmp.Diff(string(gotYAML), string(wantYAML)); diff != "" {
				t.Errorf("given %v wanted %v but got %v\n diff: %s", tt.given, want, got, diff)
				return
			}

		})

	}

}
