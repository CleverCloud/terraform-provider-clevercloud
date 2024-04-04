package helper

import "testing"

func TestRessource_String(t *testing.T) {
	tests := []struct {
		name   string
		fields *Ressource
		want   string
	}{
		// TODO: Add test cases.

		{name: "test1", fields: NewRessource("clevercloud").SetName("test1"), want: `ressource "clevercloud" "test1" {
	name = "test1"
	region = "par"
	smallest_flavor = "XS"
	biggest_flavor = "M"
	min_instance_count = 1
	max_instance_count = 2
}`},
		{name: "test2", fields: NewRessource("clevercloud_python").SetName("test2").SetIntValues("min_instance_count", 3), want: `ressource "clevercloud_python" "test2" {
	name = "test2"
	region = "par"
	smallest_flavor = "XS"
	biggest_flavor = "M"
	min_instance_count = 3
	max_instance_count = 2
}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.fields.String(); got != tt.want {
				t.Errorf("Ressource.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
