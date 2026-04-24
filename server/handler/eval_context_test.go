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

package handler

import (
	"context"
	"testing"

	"github.com/palantir/policy-bot/policy/common"
	"github.com/palantir/policy-bot/pull"
	"github.com/palantir/policy-bot/pull/pulltest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubEvaluator struct {
	result common.Result
}

func (s *stubEvaluator) Trigger() common.Trigger                                  { return common.TriggerStatic }
func (s *stubEvaluator) Evaluate(context.Context, pull.Context) common.Result    { return s.result }

func TestEvaluatePolicy_StatusMapping(t *testing.T) {
	const (
		pending  = "pending"
		failure  = "failure"
		errState = "error"
		success  = "success"
	)

	cases := []struct {
		name               string
		status             common.EvaluationStatus
		pendingStatusState string
		expected           string
	}{
		{"approved maps to success", common.StatusApproved, pending, success},
		{"disapproved maps to failure", common.StatusDisapproved, pending, failure},
		{"disapproved ignores PendingStatusState", common.StatusDisapproved, failure, failure},
		{"skipped maps to error", common.StatusSkipped, pending, errState},
		{"skipped ignores PendingStatusState", common.StatusSkipped, failure, errState},
		{"pending with default reports pending", common.StatusPending, pending, pending},
		{"pending configured as failure reports failure", common.StatusPending, failure, failure},
		{"pending configured as error reports error", common.StatusPending, errState, errState},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ec := &EvalContext{
				Options: &PullEvaluationOptions{
					StatusCheckContext: DefaultStatusCheckContext,
					PendingStatusState: tc.pendingStatusState,
				},
				PullContext: &pulltest.Context{
					OwnerValue:   "octo",
					RepoValue:    "policy-bot",
					HeadSHAValue: "deadbeef",
					BranchBaseName: "main",
					BranchHeadName: "feat",
				},
				SkipPostStatus: true,
			}

			_, err := ec.EvaluatePolicy(context.Background(), &stubEvaluator{
				result: common.Result{Status: tc.status},
			})
			require.NoError(t, err)
			require.NotNil(t, ec.Status, "expected a captured status")
			assert.Equal(t, tc.expected, ec.Status.GetState())
		})
	}
}
