package pkg

import (
	"reflect"
	"testing"
)

func TestMergeMap(t *testing.T) {
	type args struct {
		maps []map[string]string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{{
		name: "main",
		args: args{
			maps: []map[string]string{{
				"C": "D",
			}, {
				"A": "B",
			}},
		},
		want: map[string]string{
			"A": "B",
			"C": "D",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeMap(tt.args.maps...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
