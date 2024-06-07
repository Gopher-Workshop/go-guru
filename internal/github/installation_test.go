package github

import (
	"reflect"
	"testing"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

func TestInstallations_Client(t *testing.T) {
	type fields struct {
		src oauth2.TokenSource
	}
	type args struct {
		id int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "create client for installation",
			fields: fields{
				src: oauth2.ReuseTokenSource(nil, nil),
			},
			args: args{
				id: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i := NewInstallations(tt.fields.src)
			if got := i.Client(tt.args.id); got == nil {
				t.Errorf("Installations.Client() = %v, want %v", got, reflect.TypeOf(&github.Client{}))
			}
		})
	}
}
