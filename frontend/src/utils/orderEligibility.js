export function canProcessWalletRefund(order) {
  return (
    String(order?.order_status || '').trim().toLowerCase() === 'cancelled' &&
    String(order?.payment_status || '').trim().toLowerCase() === 'paid'
  );
}

export function canMarkCodPaymentPaid(order) {
  const paymentMethod = String(order?.payment_method || '').trim().toLowerCase();
  const orderStatus = String(order?.order_status || '').trim().toLowerCase();
  return (
    ['cod', 'cash_on_delivery'].includes(paymentMethod) &&
    String(order?.payment_status || '').trim().toLowerCase() === 'pending' &&
    !['cancelled', 'returned'].includes(orderStatus)
  );
}
