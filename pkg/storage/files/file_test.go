package files

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetExt(t *testing.T) {
	type args struct {
		fileName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty string",
			args: args{""},
			want: "",
		},
		{
			name: "no extension",
			args: args{"file"},
			want: "",
		},
		{
			name: "single extension",
			args: args{"file.txt"},
			want: ".txt",
		},
		{
			name: "multiple extensions",
			args: args{"file.tar.gz"},
			want: ".gz",
		},
		{
			name: "dot file",
			args: args{".bashrc"},
			want: ".bashrc",
		},
		{
			name: "complex file names",
			args: args{"example.123.abc"},
			want: ".abc",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, GetExt(tt.args.fileName), "GetExt(%v)", tt.args.fileName)
		})
	}
}
