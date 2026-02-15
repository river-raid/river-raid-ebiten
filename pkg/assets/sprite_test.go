package assets

import (
	"testing"
)

func TestNewSprite_Dimensions(t *testing.T) {
	tests := []struct {
		name                  string
		data                  []byte
		w                     int
		wantW, wantH, wantBPR int
	}{
		{
			name:    "8px wide",
			data:    make([]byte, 8), // 8 bytes for 8x8 sprite (1 byte per row)
			w:       8,
			wantW:   8,
			wantH:   8,
			wantBPR: 1,
		},
		{
			name:    "10px wide",
			data:    make([]byte, 16), // 16 bytes for 10x8 sprite (2 bytes per row)
			w:       10,
			wantW:   10,
			wantH:   8,
			wantBPR: 2,
		},
		{
			name:    "18px wide",
			data:    make([]byte, 24), // 24 bytes for 18x8 sprite (3 bytes per row)
			w:       18,
			wantW:   18,
			wantH:   8,
			wantBPR: 3,
		},
		{
			name:    "2px wide",
			data:    make([]byte, 8), // 8 bytes for 2x8 sprite (1 byte per row)
			w:       2,
			wantW:   2,
			wantH:   8,
			wantBPR: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := newSprite(tt.data, tt.w)
			if s.Width != tt.wantW {
				t.Errorf("Width = %d, want %d", s.Width, tt.wantW)
			}
			if s.Height() != tt.wantH {
				t.Errorf("Height() = %d, want %d", s.Height(), tt.wantH)
			}
			if s.BytesPerRow != tt.wantBPR {
				t.Errorf("BytesPerRow = %d, want %d", s.BytesPerRow, tt.wantBPR)
			}
		})
	}
}
