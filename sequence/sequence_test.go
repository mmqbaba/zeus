package sequence

import "testing"

func TestGetUUID(t *testing.T) {
	tests := []struct {
		name    string
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUUID()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
