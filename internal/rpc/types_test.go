// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0
package rpc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLedgerEntryResult_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		payload  string
		expected LedgerEntryResult
		wantErr  bool
	}{
		{
			name: "Valid Payload",
			payload: `{
				"key": "base64key==",
				"xdr": "base64xdr==",
				"lastModifiedLedgerSeq": 12345,
				"liveUntilLedgerSeq": 67890
			}`,
			expected: LedgerEntryResult{
				Key:                "base64key==",
				Xdr:                "base64xdr==",
				LastModifiedLedger: 12345,
				LiveUntilLedger:    67890,
			},
			wantErr: false,
		},
		{
			name: "Missing Optional Fields",
			payload: `{
				"key": "base64key==",
				"xdr": "base64xdr=="
			}`,
			expected: LedgerEntryResult{
				Key: "base64key==",
				Xdr: "base64xdr==",
			},
			wantErr: false,
		},
		{
			name:    "Invalid JSON",
			payload: `{ invalid json }`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result LedgerEntryResult
			err := json.Unmarshal([]byte(tt.payload), &result)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
