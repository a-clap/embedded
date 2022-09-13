package ds18b20

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_removeDuplicates(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "remove one duplicate",
			args: args{s: []string{"adam", "adam"}},
			want: []string{"adam"},
		},
		{
			name: "remove n duplicates",
			args: args{s: []string{"adam", "adam1", "adam2", "adam", "adam1", "adam2"}},
			want: []string{"adam", "adam1", "adam2"},
		},
		{
			name: "dont touch slice",
			args: args{s: []string{"1", "2", "3"}},
			want: []string{"1", "2", "3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := removeDuplicates(tt.args.s)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSensor_ParseTemperature(t *testing.T) {
	type fields struct {
		ID          string
		Temperature string
	}
	tests := []struct {
		name    string
		fields  fields
		want    float32
		wantErr bool
	}{
		{
			name: "five digit temperature",
			fields: fields{
				ID:          "5",
				Temperature: "12345",
			},
			want:    12.345,
			wantErr: false,
		},
		{
			name: "4 digit temperature",
			fields: fields{
				ID:          "4",
				Temperature: "1234",
			},
			want:    1.234,
			wantErr: false,
		},
		{
			name: "3 digit temperature",
			fields: fields{
				ID:          "3",
				Temperature: "234",
			},
			want:    0.234,
			wantErr: false,
		},
		{
			name: "2 digit temperature",
			fields: fields{
				ID:          "2",
				Temperature: "23",
			},
			want:    0.023,
			wantErr: false,
		},
		{
			name: "1 digit temperature",
			fields: fields{
				ID:          "1",
				Temperature: "3",
			},
			want:    0.003,
			wantErr: false,
		},
		{
			name: "not a number",
			fields: fields{
				ID:          "not a number",
				Temperature: "hey",
			},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Sensor{
				ID:          tt.fields.ID,
				Temperature: tt.fields.Temperature,
			}
			got, err := s.ParseTemperature()
			if tt.wantErr {
				assert.NotNil(t, err)
				return
			}
			require.Nil(t, err)
			assert.Equalf(t, tt.want, got, "ParseTemperature()")
		})
	}
}
