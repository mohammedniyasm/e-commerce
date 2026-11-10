package return_request

import "testing"

func TestIsValidReturnStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "requested status",
			status: ReturnStatusRequested,
			want:   true,
		},
		{
			name:   "approved status",
			status: ReturnStatusApproved,
			want:   true,
		},
		{
			name:   "rejected status",
			status: ReturnStatusRejected,
			want:   true,
		},
		{
			name:   "pickup scheduled status",
			status: ReturnStatusPickupScheduled,
			want:   true,
		},
		{
			name:   "in transit status",
			status: ReturnStatusInTransit,
			want:   true,
		},
		{
			name:   "received status",
			status: ReturnStatusReceived,
			want:   true,
		},
		{
			name:   "inspection pending status",
			status: ReturnStatusInspectionPending,
			want:   true,
		},
		{
			name:   "inspection passed status",
			status: ReturnStatusInspectionPassed,
			want:   true,
		},
		{
			name:   "inspection failed status",
			status: ReturnStatusInspectionFailed,
			want:   true,
		},
		{
			name:   "closed status",
			status: ReturnStatusClosed,
			want:   true,
		},

		// Invalid statuses
		{
			name:   "unknown status",
			status: "unknown",
			want:   false,
		},
		{
			name:   "empty status",
			status: "",
			want:   false,
		},
		{
			name:   "incorrect status casing",
			status: "REQUESTED",
			want:   false,
		},
		{
			name:   "status with whitespace",
			status: " requested ",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidReturnStatus(tt.status)

			if got != tt.want {
				t.Errorf(
					"isValidReturnStatus(%q) = %v; want %v",
					tt.status,
					got,
					tt.want,
				)
			}
		})
	}
}
func TestIsValidReturnTransition(t *testing.T) {
	tests := []struct {
		name    string
		current string
		next    string
		want    bool
	}{
		// Requested transitions
		{
			name:    "requested to approved",
			current: ReturnStatusRequested,
			next:    ReturnStatusApproved,
			want:    true,
		},
		{
			name:    "requested to rejected",
			current: ReturnStatusRequested,
			next:    ReturnStatusRejected,
			want:    true,
		},

		// Approved transitions
		{
			name:    "approved to pickup scheduled",
			current: ReturnStatusApproved,
			next:    ReturnStatusPickupScheduled,
			want:    true,
		},
		{
			name:    "approved to closed",
			current: ReturnStatusApproved,
			next:    ReturnStatusClosed,
			want:    true,
		},

		// Pickup scheduled transitions
		{
			name:    "pickup scheduled to in transit",
			current: ReturnStatusPickupScheduled,
			next:    ReturnStatusInTransit,
			want:    true,
		},
		{
			name:    "pickup scheduled to closed",
			current: ReturnStatusPickupScheduled,
			next:    ReturnStatusClosed,
			want:    true,
		},

		// In transit transitions
		{
			name:    "in transit to received",
			current: ReturnStatusInTransit,
			next:    ReturnStatusReceived,
			want:    true,
		},

		// Received transitions
		{
			name:    "received to inspection pending",
			current: ReturnStatusReceived,
			next:    ReturnStatusInspectionPending,
			want:    true,
		},

		// Inspection pending transitions
		{
			name:    "inspection pending to inspection passed",
			current: ReturnStatusInspectionPending,
			next:    ReturnStatusInspectionPassed,
			want:    true,
		},
		{
			name:    "inspection pending to inspection failed",
			current: ReturnStatusInspectionPending,
			next:    ReturnStatusInspectionFailed,
			want:    true,
		},

		// Inspection result transitions
		{
			name:    "inspection passed to closed",
			current: ReturnStatusInspectionPassed,
			next:    ReturnStatusClosed,
			want:    true,
		},
		{
			name:    "inspection failed to closed",
			current: ReturnStatusInspectionFailed,
			next:    ReturnStatusClosed,
			want:    true,
		},

		// Invalid transitions
		{
			name:    "requested to received",
			current: ReturnStatusRequested,
			next:    ReturnStatusReceived,
			want:    false,
		},
		{
			name:    "requested to closed",
			current: ReturnStatusRequested,
			next:    ReturnStatusClosed,
			want:    false,
		},
		{
			name:    "approved to rejected",
			current: ReturnStatusApproved,
			next:    ReturnStatusRejected,
			want:    false,
		},
		{
			name:    "in transit to closed",
			current: ReturnStatusInTransit,
			next:    ReturnStatusClosed,
			want:    false,
		},
		{
			name:    "received to inspection passed",
			current: ReturnStatusReceived,
			next:    ReturnStatusInspectionPassed,
			want:    false,
		},
		{
			name:    "inspection passed to inspection failed",
			current: ReturnStatusInspectionPassed,
			next:    ReturnStatusInspectionFailed,
			want:    false,
		},

		// Edge cases
		{
			name:    "same status transition",
			current: ReturnStatusRequested,
			next:    ReturnStatusRequested,
			want:    false,
		},
		{
			name:    "unknown current status",
			current: "unknown",
			next:    ReturnStatusApproved,
			want:    false,
		},
		{
			name:    "unknown next status",
			current: ReturnStatusRequested,
			next:    "unknown",
			want:    false,
		},
		{
			name:    "both statuses empty",
			current: "",
			next:    "",
			want:    false,
		},
		{
			name:    "empty current status",
			current: "",
			next:    ReturnStatusApproved,
			want:    false,
		},
		{
			name:    "empty next status",
			current: ReturnStatusRequested,
			next:    "",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidReturnTransition(tt.current, tt.next)

			if got != tt.want {
				t.Errorf(
					"isValidReturnTransition(%q, %q) = %v; want %v",
					tt.current,
					tt.next,
					got,
					tt.want,
				)
			}
		})
	}
}
