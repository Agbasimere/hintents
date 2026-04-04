// Copyright 2026 Erst Users
// SPDX-License-Identifier: Apache-2.0

// Package bridge wires snapshot compression into the IPC request pipeline.
// CompressRequest replaces the plain ledger_entries map with a Zstd-compressed,
// base64-encoded blob in ledger_entries_zstd so the Rust simulator can detect
// and decompress it automatically.
package bridge

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// ipcRequest is a minimal view of the simulator.SimulationRequest used for
// compression surgery without importing the simulator package (avoids cycles).
type ipcRequest struct {
	LedgerEntries     map[string]string `json:"ledger_entries,omitempty"`
	LedgerEntriesZstd string            `json:"ledger_entries_zstd,omitempty"`
	ControlCommand    string            `json:"control_command,omitempty"`
	RewindStep        *int              `json:"rewind_step,omitempty"`
	ForkParams        map[string]string `json:"fork_params,omitempty"`
	HarnessReset      bool              `json:"harness_reset,omitempty"`
}

const (
	// CommandRollbackAndResume requests simulator rollback to rewind_step and
	// immediate resumed execution using optional fork parameters.
	CommandRollbackAndResume = "ROLLBACK_AND_RESUME"
)

// CompressRequest takes the raw JSON bytes of a SimulationRequest, compresses
// the ledger_entries map with Zstd, and returns updated JSON bytes.
// If ledger_entries is absent or empty the input is returned unchanged.
func CompressRequest(reqJSON []byte) ([]byte, error) {
	// Unmarshal only the fields we care about.
	var partial ipcRequest
	if err := json.Unmarshal(reqJSON, &partial); err != nil {
		return nil, fmt.Errorf("bridge: unmarshal for compression: %w", err)
	}

	if len(partial.LedgerEntries) == 0 {
		return reqJSON, nil
	}

	compressed, err := CompressLedgerEntries(partial.LedgerEntries)
	if err != nil {
		return nil, err
	}

	// Patch the raw JSON: remove ledger_entries, inject ledger_entries_zstd.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(reqJSON, &raw); err != nil {
		return nil, fmt.Errorf("bridge: unmarshal raw map: %w", err)
	}

	delete(raw, "ledger_entries")

	encoded, err := json.Marshal(base64.StdEncoding.EncodeToString(compressed))
	if err != nil {
		return nil, fmt.Errorf("bridge: marshal zstd field: %w", err)
	}
	raw["ledger_entries_zstd"] = encoded

	out, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("bridge: re-marshal compressed request: %w", err)
	}
	return out, nil
}

// WithRollbackAndResume injects a rollback-and-resume control command into a
// simulation request JSON payload.
func WithRollbackAndResume(reqJSON []byte, rewindStep int, forkParams map[string]string, harnessReset bool) ([]byte, error) {
	if rewindStep < 0 {
		return nil, fmt.Errorf("bridge: rewind step must be >= 0")
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(reqJSON, &raw); err != nil {
		return nil, fmt.Errorf("bridge: unmarshal raw map: %w", err)
	}

	commandJSON, err := json.Marshal(CommandRollbackAndResume)
	if err != nil {
		return nil, fmt.Errorf("bridge: marshal control_command: %w", err)
	}
	rewindJSON, err := json.Marshal(rewindStep)
	if err != nil {
		return nil, fmt.Errorf("bridge: marshal rewind_step: %w", err)
	}

	raw["control_command"] = commandJSON
	raw["rewind_step"] = rewindJSON

	if len(forkParams) > 0 {
		paramsJSON, marshalErr := json.Marshal(forkParams)
		if marshalErr != nil {
			return nil, fmt.Errorf("bridge: marshal fork_params: %w", marshalErr)
		}
		raw["fork_params"] = paramsJSON
	} else {
		delete(raw, "fork_params")
	}

	harnessJSON, err := json.Marshal(harnessReset)
	if err != nil {
		return nil, fmt.Errorf("bridge: marshal harness_reset: %w", err)
	}
	raw["harness_reset"] = harnessJSON

	out, err := json.Marshal(raw)
	if err != nil {
		return nil, fmt.Errorf("bridge: re-marshal rollback request: %w", err)
	}

	return out, nil
}
