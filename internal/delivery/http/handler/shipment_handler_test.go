package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce/internal/delivery/http/dto/response"
	"ecommerce/internal/domain/models"
	"ecommerce/internal/mocks"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListShipmentsAppliesFiltersAndPagination(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
	shipmentUseCase := mocks.NewMockShipmentUseCase(t)
	shipmentUseCase.On(
		"ListShipments",
		mock.Anything,
		mock.MatchedBy(func(filter models.ShipmentListFilter) bool {
			return filter.Search == "BD-123" &&
				filter.ShippingStatus != nil && *filter.ShippingStatus == "in_transit" &&
				filter.Type != nil && *filter.Type == models.ShipmentTypeDelivery &&
				filter.CourierName == "Blue Dart" &&
				filter.OrderID != nil && *filter.OrderID == orderID &&
				filter.ReturnRequestID != nil && *filter.ReturnRequestID == 7 &&
				filter.CreatedFrom != nil && filter.CreatedFrom.Format("2006-01-02") == "2026-10-01" &&
				filter.CreatedTo != nil && filter.CreatedTo.Format("2006-01-02") == "2026-10-31" &&
				filter.Sort == "oldest" && filter.Page == 2 && filter.Limit == 10
		}),
	).Return([]models.Shipment{
		{
			ID:      12,
			OrderID: orderID,
			Type:    models.ShipmentTypeDelivery,
			Order:   models.Order{OrderNumber: "ORD-012"},
		},
	}, int64(21), nil).Once()

	handler := NewShipmentHandler(shipmentUseCase, nil)
	router := gin.New()
	router.GET("/admin/shipments", handler.ListShipments)
	request := httptest.NewRequest(
		http.MethodGet,
		"/admin/shipments?search=BD-123&shipping_status=in_transit&type=delivery&courier_name=Blue%20Dart&order_id="+orderID.String()+"&return_request_id=7&created_from=2026-10-01&created_to=2026-10-31&sort=oldest&page=2&limit=10",
		nil,
	)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body struct {
		Success bool                          `json:"success"`
		Data    response.ShipmentListResponse `json:"data"`
	}
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.True(t, body.Success)
	assert.Equal(t, int64(21), body.Data.TotalCount)
	assert.Equal(t, int64(3), body.Data.TotalPages)
	assert.Equal(t, 2, body.Data.Page)
	assert.Equal(t, 10, body.Data.Limit)
	require.Len(t, body.Data.Shipments, 1)
	assert.Equal(t, "ORD-012", body.Data.Shipments[0].OrderNumber)
}

func TestListShipmentsRejectsInvalidFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	shipmentUseCase := mocks.NewMockShipmentUseCase(t)
	handler := NewShipmentHandler(shipmentUseCase, nil)
	router := gin.New()
	router.GET("/admin/shipments", handler.ListShipments)

	for _, query := range []string{
		"shipping_status=unknown",
		"type=unknown",
		"order_id=not-a-uuid",
		"return_request_id=0",
		"created_from=2026-10-31&created_to=2026-10-01",
		"sort=untrusted",
	} {
		t.Run(query, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/admin/shipments?"+query, nil)
			router.ServeHTTP(recorder, request)
			assert.Equal(t, http.StatusBadRequest, recorder.Code)
		})
	}
}
