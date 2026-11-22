package models

import (
	"encoding/json"
	"testing"
)

func TestDashboardOverviewUsesFrontendJSONFieldNames(t *testing.T) {
	overview := DashboardOverview{
		TotalRevenue: 1200,
		TotalOrders:  12,
	}

	data, err := json.Marshal(overview)
	if err != nil {
		t.Fatalf("marshal dashboard overview: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("unmarshal dashboard overview: %v", err)
	}

	if fields["total_orders"] != float64(12) {
		t.Errorf("total_orders = %v, want 12", fields["total_orders"])
	}
	if fields["total_revenue"] != float64(1200) {
		t.Errorf("total_revenue = %v, want 1200", fields["total_revenue"])
	}
	if _, exists := fields["TotalOrders"]; exists {
		t.Error("dashboard overview should not expose Go field names")
	}
}
