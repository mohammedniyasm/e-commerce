package order

import "testing"

func TestIsValidOrderTransition(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		nextStatus    string
		want          bool
	}{
		// Valid transitions
		{
			name:          "pending to confirmed",
			currentStatus: "pending",
			nextStatus:    "confirmed",
			want:          true,
		},
		{
			name:          "pending to cancelled",
			currentStatus: "pending",
			nextStatus:    "cancelled",
			want:          true,
		},
		{
			name:          "confirmed to processing",
			currentStatus: "confirmed",
			nextStatus:    "processing",
			want:          true,
		},
		{
			name:          "confirmed to cancelled",
			currentStatus: "confirmed",
			nextStatus:    "cancelled",
			want:          true,
		},
		{
			name:          "processing to shipped",
			currentStatus: "processing",
			nextStatus:    "shipped",
			want:          true,
		},
		{
			name:          "processing to cancelled",
			currentStatus: "processing",
			nextStatus:    "cancelled",
			want:          true,
		},
		{
			name:          "shipped to delivered",
			currentStatus: "shipped",
			nextStatus:    "delivered",
			want:          true,
		},
		{
			name:          "shipped to returned",
			currentStatus: "shipped",
			nextStatus:    "returned",
			want:          true,
		},
		{
			name:          "out for delivery to delivered",
			currentStatus: "out_for_delivery",
			nextStatus:    "delivered",
			want:          true,
		},
		{
			name:          "out for delivery to cancelled",
			currentStatus: "out_for_delivery",
			nextStatus:    "cancelled",
			want:          true,
		},
		{
			name:          "delivered to returned",
			currentStatus: "delivered",
			nextStatus:    "returned",
			want:          true,
		},
		{
			name:          "delivered to processing",
			currentStatus: "delivered",
			nextStatus:    "processing",
			want:          true,
		},

		// Invalid transitions
		{
			name:          "pending to shipped",
			currentStatus: "pending",
			nextStatus:    "shipped",
			want:          false,
		},
		{
			name:          "confirmed to delivered",
			currentStatus: "confirmed",
			nextStatus:    "delivered",
			want:          false,
		},
		{
			name:          "processing to pending",
			currentStatus: "processing",
			nextStatus:    "pending",
			want:          false,
		},
		{
			name:          "shipped to cancelled",
			currentStatus: "shipped",
			nextStatus:    "cancelled",
			want:          false,
		},
		{
			name:          "delivered to cancelled",
			currentStatus: "delivered",
			nextStatus:    "cancelled",
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
			nextStatus:    "confirmed",
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
			nextStatus:    "confirmed",
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
			got := isValidOrderTransition(tt.currentStatus, tt.nextStatus)

			if got != tt.want {
				t.Errorf(
					"isValidOrderTransition(%q, %q) = %v; want %v",
					tt.currentStatus,
					tt.nextStatus,
					got,
					tt.want,
				)
			}
		})
	}
}
