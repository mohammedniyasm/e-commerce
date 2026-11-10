package shipment

import "testing"

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
func TestIsValidReturnStatus(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   bool
	}{
		{
			name:   "requested",
			status: ReturnStatusRequested,
			want:   true,
		},
		{
			name:   "approved",
			status: ReturnStatusApproved,
			want:   true,
		},
		{
			name:   "rejected",
			status: ReturnStatusRejected,
			want:   true,
		},
		{
			name:   "pickup scheduled",
			status: ReturnStatusPickupScheduled,
			want:   true,
		},
		{
			name:   "in transit",
			status: ReturnStatusInTransit,
			want:   true,
		},
		{
			name:   "received",
			status: ReturnStatusReceived,
			want:   true,
		},
		{
			name:   "inspection pending",
			status: ReturnStatusInspectionPending,
			want:   true,
		},
		{
			name:   "inspection passed",
			status: ReturnStatusInspectionPassed,
			want:   true,
		},
		{
			name:   "inspection failed",
			status: ReturnStatusInspectionFailed,
			want:   true,
		},
		{
			name:   "closed",
			status: ReturnStatusClosed,
			want:   true,
		},
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
			name:   "incorrect casing",
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
				t.Errorf("isValidReturnStatus(%q) = %v; want %v",
					tt.status, got, tt.want)
			}
		})
	}
}
func TestIsValidShippingTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		nextStatus    string
		want          bool
	}{
		// Pending transitions
		{
			name:          "pending to packed",
			currentStatus: "pending",
			nextStatus:    "packed",
			want:          true,
		},
		{
			name:          "pending to cancelled",
			currentStatus: "pending",
			nextStatus:    "cancelled",
			want:          true,
		},

		// Packed transitions
		{
			name:          "packed to ready to ship",
			currentStatus: "packed",
			nextStatus:    "ready_to_ship",
			want:          true,
		},
		{
			name:          "packed to cancelled",
			currentStatus: "packed",
			nextStatus:    "cancelled",
			want:          true,
		},

		// Ready to ship transitions
		{
			name:          "ready to ship to in transit",
			currentStatus: "ready_to_ship",
			nextStatus:    "in_transit",
			want:          true,
		},
		{
			name:          "ready to ship to cancelled",
			currentStatus: "ready_to_ship",
			nextStatus:    "cancelled",
			want:          true,
		},

		// In transit transitions
		{
			name:          "in transit to out for delivery",
			currentStatus: "in_transit",
			nextStatus:    "out_for_delivery",
			want:          true,
		},
		{
			name:          "in transit to failed",
			currentStatus: "in_transit",
			nextStatus:    "failed",
			want:          true,
		},
		{
			name:          "in transit to returned",
			currentStatus: "in_transit",
			nextStatus:    "returned",
			want:          true,
		},

		// Out for delivery transitions
		{
			name:          "out for delivery to delivered",
			currentStatus: "out_for_delivery",
			nextStatus:    "delivered",
			want:          true,
		},
		{
			name:          "out for delivery to failed",
			currentStatus: "out_for_delivery",
			nextStatus:    "failed",
			want:          true,
		},
		{
			name:          "out for delivery to returned",
			currentStatus: "out_for_delivery",
			nextStatus:    "returned",
			want:          true,
		},

		// Failed transitions
		{
			name:          "failed to in transit",
			currentStatus: "failed",
			nextStatus:    "in_transit",
			want:          true,
		},
		{
			name:          "failed to out for delivery",
			currentStatus: "failed",
			nextStatus:    "out_for_delivery",
			want:          true,
		},
		{
			name:          "failed to returned",
			currentStatus: "failed",
			nextStatus:    "returned",
			want:          true,
		},

		// Invalid transitions
		{
			name:          "pending to delivered",
			currentStatus: "pending",
			nextStatus:    "delivered",
			want:          false,
		},
		{
			name:          "packed to in transit",
			currentStatus: "packed",
			nextStatus:    "in_transit",
			want:          false,
		},
		{
			name:          "ready to ship to delivered",
			currentStatus: "ready_to_ship",
			nextStatus:    "delivered",
			want:          false,
		},
		{
			name:          "in transit to cancelled",
			currentStatus: "in_transit",
			nextStatus:    "cancelled",
			want:          false,
		},
		{
			name:          "out for delivery to cancelled",
			currentStatus: "out_for_delivery",
			nextStatus:    "cancelled",
			want:          false,
		},
		{
			name:          "failed to cancelled",
			currentStatus: "failed",
			nextStatus:    "cancelled",
			want:          false,
		},
		{
			name:          "failed to delivered",
			currentStatus: "failed",
			nextStatus:    "delivered",
			want:          false,
		},

		// Edge cases
		{
			name:          "same status transition",
			currentStatus: "pending",
			nextStatus:    "pending",
			want:          false,
		},
		{
			name:          "unknown current status",
			currentStatus: "unknown",
			nextStatus:    "packed",
			want:          false,
		},
		{
			name:          "unknown next status",
			currentStatus: "pending",
			nextStatus:    "unknown",
			want:          false,
		},
		{
			name:          "empty current status",
			currentStatus: "",
			nextStatus:    "packed",
			want:          false,
		},
		{
			name:          "empty next status",
			currentStatus: "pending",
			nextStatus:    "",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidShippingTransition(tt.currentStatus, tt.nextStatus)

			if got != tt.want {
				t.Errorf(
					"isValidShippingTransition(%q, %q) = %v; want %v",
					tt.currentStatus,
					tt.nextStatus,
					got,
					tt.want,
				)
			}
		})
	}
}
