package cases

import (
	"testing"
)

func TestBoolConstFalse_ValidateAll(t *testing.T) {
	type fields struct {
		Val bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"test", fields{Val: false}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BoolConstFalse{
				Val: tt.fields.Val,
			}
			if err := m.ValidateAll(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBoolConstTrue_ValidateAll(t *testing.T) {
	type fields struct {
		Val bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			"test", fields{Val: true}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &BoolConstTrue{
				Val: tt.fields.Val,
			}
			if err := m.ValidateAll(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
