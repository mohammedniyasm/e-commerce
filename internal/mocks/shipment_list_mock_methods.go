package mocks

import (
	"context"

	"ecommerce/internal/domain/models"
)

func (_mock *MockShipmentRepository) List(ctx context.Context, filter models.ShipmentListFilter) ([]models.Shipment, int64, error) {
	ret := _mock.Called(ctx, filter)
	if len(ret) == 0 {
		panic("no return value specified for List")
	}

	var shipments []models.Shipment
	if returnFunc, ok := ret.Get(0).(func(context.Context, models.ShipmentListFilter) ([]models.Shipment, int64, error)); ok {
		return returnFunc(ctx, filter)
	}
	if ret.Get(0) != nil {
		shipments = ret.Get(0).([]models.Shipment)
	}

	var total int64
	if returnFunc, ok := ret.Get(1).(func(context.Context, models.ShipmentListFilter) int64); ok {
		total = returnFunc(ctx, filter)
	} else if ret.Get(1) != nil {
		total = ret.Get(1).(int64)
	}
	return shipments, total, ret.Error(2)
}

func (_mock *MockShipmentUseCase) ListShipments(ctx context.Context, filter models.ShipmentListFilter) ([]models.Shipment, int64, error) {
	ret := _mock.Called(ctx, filter)
	if len(ret) == 0 {
		panic("no return value specified for ListShipments")
	}

	var shipments []models.Shipment
	if returnFunc, ok := ret.Get(0).(func(context.Context, models.ShipmentListFilter) ([]models.Shipment, int64, error)); ok {
		return returnFunc(ctx, filter)
	}
	if ret.Get(0) != nil {
		shipments = ret.Get(0).([]models.Shipment)
	}

	var total int64
	if returnFunc, ok := ret.Get(1).(func(context.Context, models.ShipmentListFilter) int64); ok {
		total = returnFunc(ctx, filter)
	} else if ret.Get(1) != nil {
		total = ret.Get(1).(int64)
	}
	return shipments, total, ret.Error(2)
}
