package external

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateKey(t *testing.T) {
	tests := []struct {
		name     string
		orgID    string
		folder   string
		filename string
		wantPrefix string
	}{
		{
			name:       "Standard image",
			orgID:      "org-123",
			folder:     "references",
			filename:   "image.jpg",
			wantPrefix: "org-123/references/",
		},
		{
			name:       "Product image",
			orgID:      "org-456",
			folder:     "products",
			filename:   "product.png",
			wantPrefix: "org-456/products/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := GenerateKey(tt.orgID, tt.folder, tt.filename)
			
			assert.True(t, strings.HasPrefix(key, tt.wantPrefix))
			assert.True(t, strings.HasSuffix(key, tt.filename))
			assert.Contains(t, key, "_") // Should have timestamp separator
		})
	}
}


