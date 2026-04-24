// Copyright 2026 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseConfig_PendingStatusState(t *testing.T) {
	// Isolate from any ambient env var that might override the YAML value.
	t.Setenv("POLICYBOT_ENV_PREFIX", "POLICYBOT_TEST_UNSET_")

	t.Run("default when field omitted", func(t *testing.T) {
		c, err := ParseConfig([]byte(`options: {}`))
		require.NoError(t, err)
		assert.Equal(t, "pending", c.Options.PendingStatusState)
	})

	t.Run("default when field empty string", func(t *testing.T) {
		c, err := ParseConfig([]byte("options:\n  pending_status_state: \"\"\n"))
		require.NoError(t, err)
		assert.Equal(t, "pending", c.Options.PendingStatusState)
	})

	for _, v := range []string{"pending", "failure", "error"} {
		t.Run("allows "+v, func(t *testing.T) {
			c, err := ParseConfig([]byte("options:\n  pending_status_state: " + v + "\n"))
			require.NoError(t, err)
			assert.Equal(t, v, c.Options.PendingStatusState)
		})
	}

	t.Run("rejects success", func(t *testing.T) {
		_, err := ParseConfig([]byte("options:\n  pending_status_state: success\n"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pending_status_state")
	})

	t.Run("rejects garbage", func(t *testing.T) {
		_, err := ParseConfig([]byte("options:\n  pending_status_state: banana\n"))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pending_status_state")
	})

	t.Run("rejects misspelled sibling key via strict unmarshal", func(t *testing.T) {
		_, err := ParseConfig([]byte("options:\n  pendng_status_state: pending\n"))
		require.Error(t, err)
		// yaml.UnmarshalStrict returns an error containing the unknown field name.
		assert.True(t,
			strings.Contains(err.Error(), "pendng_status_state") ||
				strings.Contains(err.Error(), "not found"),
			"expected strict-unmarshal error, got: %v", err,
		)
	})
}

func TestParseConfig_PendingStatusState_FromEnv(t *testing.T) {
	t.Run("env var overrides YAML and validates", func(t *testing.T) {
		t.Setenv("POLICYBOT_OPTIONS_PENDING_STATUS_STATE", "failure")
		c, err := ParseConfig([]byte("options:\n  pending_status_state: pending\n"))
		require.NoError(t, err)
		assert.Equal(t, "failure", c.Options.PendingStatusState)
	})

	t.Run("invalid env var is rejected at startup", func(t *testing.T) {
		t.Setenv("POLICYBOT_OPTIONS_PENDING_STATUS_STATE", "success")
		_, err := ParseConfig([]byte(`options: {}`))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "pending_status_state")
	})
}
