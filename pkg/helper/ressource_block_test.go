package helper

import "testing"

func TestRessource_String(t *testing.T) {
	tests := []struct {
		name   string
		fields *Ressource
		want   string
	}{
		{name: "test1",
			fields: NewRessource("clevercloud", "test1"),
			want: `resource "clevercloud" "test1" {
}
`},
		{
			name: "test2",
			fields: NewRessource("clevercloud_python", "test2").
				SetOneValue("biggest_flavor", "XXL").
				SetOneValue("test", 3),
			want: `resource "clevercloud_python" "test2" {
	biggest_flavor = "XXL"
	test = 3
}
`},
		{name: "test3",
			fields: NewRessource(
				"clevercloud_python",
				"test3",
				SetKeyValues(map[string]any{
					"region":             "ici",
					"teststring":         "smt",
					"min_instance_count": 0,
					"testint":            12,
					"map":                map[string]any{"test_string": "string", "test_int": 42}}),
				SetBlockValues("testblock", map[string]any{"test_string": "string", "test_int": 42})),
			want: `resource "clevercloud_python" "test3" {
	map = {
		test_int = 42
		test_string = "string"
	}
	min_instance_count = 0
	region = "ici"
	testint = 12
	teststring = "smt"
	testblock {
		test_int = 42
		test_string = "string"
	}
}
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.String(); got != tt.want {
				t.Errorf("Ressource.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
