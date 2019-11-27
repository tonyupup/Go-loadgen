package my

import (
	"reflect"
	"testing"
	"time"
)

func Test_myCaller_Call(t *testing.T) {
	type fields struct {
		id uint32
	}
	type args struct {
		req       []byte
		timeoutNs time.Duration
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []byte
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &myCaller{
				id: tt.fields.id,
			}
			got, err := c.Call(tt.args.req, tt.args.timeoutNs)
			if (err != nil) != tt.wantErr {
				t.Errorf("myCaller.Call() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("myCaller.Call() = %v, want %v", got, tt.want)
			}
		})
	}
}
