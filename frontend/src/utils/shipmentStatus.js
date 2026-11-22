export const SHIPMENT_STATUS_LABELS = {
  pending: 'Pending',
  packed: 'Packed',
  ready_to_ship: 'Ready to Ship',
  in_transit: 'In Transit',
  out_for_delivery: 'Out for Delivery',
  delivered: 'Delivered',
  failed: 'Failed',
  returned: 'Returned',
  cancelled: 'Cancelled',
};

export const SHIPMENT_STATUS_TRANSITIONS = {
  pending: ['packed', 'cancelled'],
  packed: ['ready_to_ship', 'cancelled'],
  ready_to_ship: ['in_transit', 'cancelled'],
  in_transit: ['out_for_delivery', 'failed', 'returned'],
  out_for_delivery: ['delivered', 'failed', 'returned'],
  failed: ['in_transit', 'out_for_delivery', 'returned'],
  delivered: [],
  returned: [],
  cancelled: [],
};

export const SHIPMENT_TYPE_LABELS = {
  delivery: 'Delivery',
  return: 'Return',
  return_to_customer: 'Return to Customer',
};

export function getShipmentStatusTransitions(status) {
  return SHIPMENT_STATUS_TRANSITIONS[status] ?? [];
}
