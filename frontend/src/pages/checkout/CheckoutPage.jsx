import { useEffect, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import {
  AlertCircle,
  AlertTriangle,
  ArrowRight,
  CheckCircle2,
  CreditCard,
  MapPin,
  Plus,
  RefreshCw,
  ShieldCheck,
  Tag,
  Truck,
  Wallet,
} from 'lucide-react';
import {
  addUserAddress,
  cancelOrder,
  createRazorpayOrder,
  getMyOrderById,
  getUserAddresses,
  getUserWallet,
  placeOrder,
  validateCoupon,
  verifyRazorpayPayment,
} from '../../api/checkoutService';
import { useAuth } from '../../context/AuthContext';
import { useCart } from '../../context/CartContext';
import { useToast } from '../../context/ToastContext';
import { formatCurrencyINR } from '../../utils/formatters';
import { Badge } from '../../components/common/Badge';
import { Button } from '../../components/common/Button';
import { FormField } from '../../components/common/FormField';
import { Input } from '../../components/common/Input';
import { Modal } from '../../components/common/Modal';
import { SkeletonLoader } from '../../components/common/SkeletonLoader';
import './CheckoutPage.css';

const PENDING_ORDER_STORAGE_KEY = 'drape_pending_order_id';

/**
 * Helper to dynamically load the Razorpay checkout script if not present.
 */
function loadRazorpayScript() {
  return new Promise((resolve) => {
    if (window.Razorpay) {
      resolve(true);
      return;
    }
    const script = document.createElement('script');
    script.src = 'https://checkout.razorpay.com/v1/checkout.js';
    script.onload = () => resolve(true);
    script.onerror = () => resolve(false);
    document.body.appendChild(script);
  });
}

export function CheckoutPage() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const { cart, clearCart, isLoading: isCartLoading } = useCart();
  const { showToast } = useToast();

  // State
  const [addresses, setAddresses] = useState([]);
  const [selectedAddressId, setSelectedAddressId] = useState(null);
  const [wallet, setWallet] = useState(null);
  const [paymentMethod, setPaymentMethod] = useState('online'); // 'online' | 'cod' | 'wallet'

  // Coupon
  const [couponCode, setCouponCode] = useState('');
  const [appliedCoupon, setAppliedCoupon] = useState(null);
  const [couponError, setCouponError] = useState('');
  const [isApplyingCoupon, setIsApplyingCoupon] = useState(false);

  // Modal for new address
  const [isAddressModalOpen, setIsAddressModalOpen] = useState(false);
  const [newAddress, setNewAddress] = useState({
    name: '',
    phone: '',
    address_line1: '',
    address_line_2: '',
    city: '',
    state: '',
    postal_code: '',
    is_default: false,
  });
  const [addressFormError, setAddressFormError] = useState('');
  const [isSavingAddress, setIsSavingAddress] = useState(false);

  // Order submission
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [createdPendingOrder, setCreatedPendingOrder] = useState(null);
  const [paymentError, setPaymentError] = useState('');
  const [isLoadingPage, setIsLoadingPage] = useState(true);

  const cartItems = cart?.items || [];
  const displayItems =
    cartItems.length > 0 ? cartItems : (createdPendingOrder?.items || []);
  const rawSubtotal =
    cartItems.length > 0
      ? Number(cart?.available_subtotal || cart?.subtotal || 0)
      : Number(createdPendingOrder?.subtotal || createdPendingOrder?.total_amount || 0);

  // Load addresses, wallet and check for recoverable pending order
  useEffect(() => {
    let isMounted = true;

    async function loadCheckoutData() {
      setIsLoadingPage(true);
      try {
        const [fetchedAddresses, fetchedWallet] = await Promise.allSettled([
          getUserAddresses(),
          getUserWallet(),
        ]);

        if (!isMounted) return;

        if (fetchedAddresses.status === 'fulfilled' && Array.isArray(fetchedAddresses.value)) {
          setAddresses(fetchedAddresses.value);
          const defaultAddr = fetchedAddresses.value.find((a) => a.is_default);
          if (defaultAddr) {
            setSelectedAddressId(defaultAddr.id);
          } else if (fetchedAddresses.value.length > 0) {
            setSelectedAddressId(fetchedAddresses.value[0].id);
          }
        }

        if (fetchedWallet.status === 'fulfilled' && fetchedWallet.value) {
          setWallet(fetchedWallet.value);
        }

        // Recover pending online order if stored in sessionStorage
        if (typeof sessionStorage !== 'undefined') {
          const storedPendingId = sessionStorage.getItem(PENDING_ORDER_STORAGE_KEY);
          if (storedPendingId) {
            try {
              const recovered = await getMyOrderById(storedPendingId);
              const orderStatus = (recovered?.order_status || '').toLowerCase();
              const paymentStatus = (recovered?.payment_status || '').toLowerCase();
              const placedTime = new Date(recovered?.placed_at || recovered?.created_at).getTime();
              // Backend order expiry deadline is 10 minutes
              const isExpired = !isNaN(placedTime) && Date.now() - placedTime > 10 * 60 * 1000;

              const isEligibleForRetry =
                orderStatus === 'pending' && paymentStatus === 'pending' && !isExpired;

              if (isMounted && recovered && isEligibleForRetry) {
                setCreatedPendingOrder({
                  ...recovered,
                  order_id: recovered.id || recovered.order_id,
                });
                setPaymentMethod('online');
                setPaymentError(
                  `Order #${recovered.order_number} has an incomplete online payment. You can resume checkout below.`,
                );
              } else if (isMounted && recovered && orderStatus === 'pending' && isExpired) {
                // Expired on the clock, but order status is still pending on backend!
                // Stock is STILL reserved on backend; do NOT silently discard ID.
                setCreatedPendingOrder({
                  ...recovered,
                  order_id: recovered.id || recovered.order_id,
                  is_expired: true,
                });
                setPaymentMethod('online');
                setPaymentError(
                  `The 10-minute payment window for Order #${recovered.order_number} has expired. Please cancel this order to release reserved items.`,
                );
              } else if (
                recovered &&
                orderStatus === 'confirmed' &&
                paymentStatus === 'failed'
              ) {
                // Webhook-confirmed terminal payment failure: non-retryable
                sessionStorage.removeItem(PENDING_ORDER_STORAGE_KEY);
                setCreatedPendingOrder(null);
                setPaymentError(
                  `Payment for Order #${recovered.order_number} could not be completed and has failed permanently. This order cannot be retried. Please review your bag and place a new order.`,
                );
                showToast({
                  title: 'Payment Failed',
                  message: `Payment for Order #${recovered.order_number} failed permanently. This order cannot be retried.`,
                  type: 'error',
                });
              } else if (
                recovered &&
                (orderStatus === 'cancelled' ||
                  orderStatus === 'expired' ||
                  paymentStatus === 'paid')
              ) {
                // Clear ONLY upon backend-confirmed terminal state
                sessionStorage.removeItem(PENDING_ORDER_STORAGE_KEY);
              }
            } catch (recoveryErr) {
              const status = recoveryErr.response?.status;
              if (status === 401) {
                // Auth expired or token refresh in-flight: DO NOT clear order ID!
                setPaymentError(
                  'Authentication session expired while verifying pending order. Please re-authenticate.',
                );
              } else if (status === 404) {
                // Confirm 404 represents genuinely unavailable order before clearing
                const msg = (recoveryErr.response?.data?.message || '').toLowerCase();
                if (msg.includes('order not found') || msg.includes('not found')) {
                  sessionStorage.removeItem(PENDING_ORDER_STORAGE_KEY);
                } else {
                  setPaymentError('Pending order could not be located on the server.');
                }
              } else {
                // Network error or server downtime: DO NOT delete stored order ID
                setPaymentError(
                  'Unable to reach server to verify pending order. Please check your connection.',
                );
              }
            }
          }
        }
      } catch (err) {
        showToast({
          title: 'Checkout Warning',
          message: 'Unable to load all profile details. You can still enter your address.',
          type: 'warning',
        });
      } finally {
        if (isMounted) setIsLoadingPage(false);
      }
    }

    loadCheckoutData();

    return () => {
      isMounted = false;
    };
  }, []);

  // Handle coupon validation
  async function handleApplyCoupon(e) {
    e.preventDefault();
    if (!couponCode.trim()) {
      setCouponError('Please enter a coupon code.');
      return;
    }
    setCouponError('');
    setIsApplyingCoupon(true);

    try {
      const result = await validateCoupon({
        code: couponCode.trim(),
        order_amount: rawSubtotal,
      });

      setAppliedCoupon(result);
      showToast({
        title: 'Coupon Applied',
        message: `Saved ${formatCurrencyINR(result.discount_amount || 0)} on this order.`,
        type: 'success',
      });
    } catch (err) {
      setCouponError(
        err.response?.data?.message || err.message || 'Invalid or inapplicable coupon code.',
      );
      setAppliedCoupon(null);
    } finally {
      setIsApplyingCoupon(false);
    }
  }

  function handleRemoveCoupon() {
    setAppliedCoupon(null);
    setCouponCode('');
    setCouponError('');
  }

  // Handle saving new address
  async function handleSaveAddress(e) {
    e.preventDefault();
    if (
      !newAddress.name.trim() ||
      !newAddress.phone.trim() ||
      !newAddress.address_line1.trim() ||
      !newAddress.city.trim() ||
      !newAddress.state.trim() ||
      !newAddress.postal_code.trim()
    ) {
      setAddressFormError('Please complete all required fields (including State).');
      return;
    }

    setAddressFormError('');
    setIsSavingAddress(true);
    try {
      const created = await addUserAddress({
        name: newAddress.name.trim(),
        phone: newAddress.phone.trim(),
        address_line1: newAddress.address_line1.trim(),
        address_line_2: newAddress.address_line_2.trim() || undefined,
        city: newAddress.city.trim(),
        state: newAddress.state.trim(),
        postal_code: newAddress.postal_code.trim(),
        is_default: Boolean(newAddress.is_default),
      });

      const updatedAddresses = await getUserAddresses();
      setAddresses(updatedAddresses);
      if (created?.id) {
        setSelectedAddressId(created.id);
      } else if (updatedAddresses.length > 0) {
        setSelectedAddressId(updatedAddresses[updatedAddresses.length - 1].id);
      }

      setIsAddressModalOpen(false);
      setNewAddress({
        name: '',
        phone: '',
        address_line1: '',
        address_line_2: '',
        city: '',
        state: '',
        postal_code: '',
        is_default: false,
      });
      showToast({
        title: 'Address Saved',
        message: 'Delivery address added successfully.',
        type: 'success',
      });
    } catch (err) {
      setAddressFormError(err.response?.data?.message || err.message || 'Failed to save address.');
    } finally {
      setIsSavingAddress(false);
    }
  }

  // Authoritative calculations
  const discountAmount = Number(appliedCoupon?.discount_amount || 0);
  const estimatedTotal = Math.max(0, rawSubtotal - discountAmount);

  const canUseWallet = wallet && Number(wallet.balance || 0) >= estimatedTotal;

  // Execute Razorpay checkout flow with backend verification
  async function executeRazorpayFlow(order) {
    setPaymentError('');
    const isScriptLoaded = await loadRazorpayScript();

    if (!isScriptLoaded || !window.Razorpay) {
      // In offline/test environments without Razorpay SDK, simulate or warn
      setPaymentError(
        'Online payment gateway is temporarily unreachable. Please retry or choose Cash on Delivery.',
      );
      return;
    }

    try {
      const razorpayOrder = await createRazorpayOrder({ order_id: order.order_id });

      const options = {
        key: import.meta.env.VITE_RAZORPAY_KEY_ID || 'rzp_test_Tvc7Cx1Mt8YYEk',
        amount: Math.round(order.total_amount * 100),
        currency: 'INR',
        name: 'Drape Atelier',
        description: `Order #${order.order_number}`,
        order_id: razorpayOrder.razorpay_order_id,
        prefill: {
          name: user?.name || '',
          email: user?.email || '',
        },
        theme: {
          color: '#1c1917',
        },
        handler: async function (response) {
          try {
            setIsSubmitting(true);
            // CRITICAL: Must verify payment on backend
            await verifyRazorpayPayment({
              order_id: order.order_id,
              razorpay_order_id: response.razorpay_order_id,
              razorpay_payment_id: response.razorpay_payment_id,
              razorpay_signature: response.razorpay_signature,
            });

            // Verification successful
            if (typeof sessionStorage !== 'undefined') {
              sessionStorage.removeItem(PENDING_ORDER_STORAGE_KEY);
            }
            await clearCart();
            navigate('/checkout/success', {
              state: {
                order: {
                  ...order,
                  payment_status: 'paid',
                },
              },
            });
          } catch (verifyErr) {
            setPaymentError(
              verifyErr.response?.data?.message ||
                'Payment verification failed. Please contact concierge support.',
            );
          } finally {
            setIsSubmitting(false);
          }
        },
        modal: {
          ondismiss: function () {
            setPaymentError(
              'Payment window was closed before completion. Your order was created as pending. You may retry payment below.',
            );
          },
        },
      };

      const rzp = new window.Razorpay(options);
      rzp.on('payment.failed', function (resp) {
        setPaymentError(
          resp.error?.description || 'Payment failed at the gateway. Please retry or use another card.',
        );
      });
      rzp.open();
    } catch (err) {
      setPaymentError(
        err.response?.data?.message || err.message || 'Failed to initialize payment gateway.',
      );
    }
  }

  // Handle Order Submission
  async function handleSubmitOrder() {
    if (!selectedAddressId) {
      showToast({
        title: 'Address Required',
        message: 'Please select or add a delivery address to proceed.',
        type: 'warning',
      });
      return;
    }

    if (cartItems.length === 0) {
      showToast({
        title: 'Cart Empty',
        message: 'Your shopping bag has no items.',
        type: 'warning',
      });
      return;
    }

    if (cart.has_unavailable_items) {
      showToast({
        title: 'Unavailable Items',
        message: 'Please remove out-of-stock items from your bag before checking out.',
        type: 'error',
      });
      return;
    }

    // Prevent duplicate orders
    if (isSubmitting) return;

    // If an order has already been created and is pending payment retry
    if (createdPendingOrder) {
      if (createdPendingOrder.is_expired) {
        showToast({
          title: 'Payment Window Expired',
          message:
            'The 10-minute payment window has expired. Please cancel this order to release reserved items.',
          type: 'warning',
        });
        return;
      }
      if (paymentMethod === 'online') {
        await executeRazorpayFlow(createdPendingOrder);
        return;
      }
      showToast({
        title: 'Pending Order Active',
        message:
          'You have a pending online order. Please complete payment or cancel it before selecting another payment method.',
        type: 'warning',
      });
      return;
    }

    setIsSubmitting(true);
    setPaymentError('');

    try {
      const cartItemIds = cartItems.map((item) => item.id).filter(Boolean);

      const placedOrder = await placeOrder({
        cart_item_ids: cartItemIds,
        address_id: selectedAddressId,
        payment_method: paymentMethod,
        coupon_code: appliedCoupon ? appliedCoupon.code : undefined,
      });

      // Handle based on payment method
      if (paymentMethod === 'cod' || paymentMethod === 'wallet') {
        await clearCart();
        navigate('/checkout/success', {
          state: {
            order: placedOrder,
          },
        });
      } else if (paymentMethod === 'online') {
        // Online Payment: Razorpay
        if (typeof sessionStorage !== 'undefined' && placedOrder?.order_id) {
          sessionStorage.setItem(PENDING_ORDER_STORAGE_KEY, placedOrder.order_id);
        }
        setCreatedPendingOrder(placedOrder);
        await executeRazorpayFlow(placedOrder);
      }
    } catch (err) {
      const errMsg =
        err.response?.data?.message || err.message || 'Unable to place order. Please review your details.';
      setPaymentError(errMsg);
      showToast({
        title: 'Order Placement Error',
        message: errMsg,
        type: 'error',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  // Handle Cancellation of Pending Order via Backend
  async function handleCancelPendingOrder() {
    if (!createdPendingOrder?.order_id) return;
    setIsSubmitting(true);
    try {
      await cancelOrder(createdPendingOrder.order_id);
      if (typeof sessionStorage !== 'undefined') {
        sessionStorage.removeItem(PENDING_ORDER_STORAGE_KEY);
      }
      setCreatedPendingOrder(null);
      setPaymentError('');
      showToast({
        title: 'Order Cancelled',
        message: 'Pending order was cancelled. You may now modify your bag or place a new order.',
        type: 'info',
      });
    } catch (err) {
      showToast({
        title: 'Cancellation Error',
        message: err.response?.data?.message || err.message || 'Unable to cancel pending order.',
        type: 'error',
      });
    } finally {
      setIsSubmitting(false);
    }
  }

  if (isLoadingPage || isCartLoading) {
    return (
      <div className="checkout-page checkout-loading" data-testid="checkout-loading">
        <SkeletonLoader lines={8} />
      </div>
    );
  }

  if (cartItems.length === 0 && !createdPendingOrder) {
    return (
      <div className="checkout-page checkout-empty" data-testid="checkout-empty">
        {paymentError && (
          <div
            className="checkout-payment-error"
            role="alert"
            data-testid="payment-error-banner"
            style={{ marginBottom: '24px', width: '100%', maxWidth: '600px' }}
          >
            <AlertTriangle size={20} />
            <div className="checkout-payment-error__content">
              <span className="checkout-payment-error__title">Payment Attention</span>
              <p className="checkout-payment-error__msg">{paymentError}</p>
            </div>
          </div>
        )}
        <AlertCircle size={48} className="checkout-empty-icon" />
        <h2>Your Bag is Empty</h2>
        <p>You cannot proceed to checkout without any items in your shopping bag.</p>
        <Link to="/catalog">
          <Button variant="primary" size="md">
            RETURN TO COLLECTION
          </Button>
        </Link>
      </div>
    );
  }

  return (
    <div className="checkout-page" data-testid="checkout-page">
      <div className="checkout-header">
        <h1 className="checkout-title">Checkout</h1>
        <p className="checkout-subtitle">Secure luxury order completion & atelier delivery</p>
      </div>

      {paymentError && (
        <div className="checkout-payment-error" role="alert" data-testid="payment-error-banner">
          <AlertTriangle size={20} />
          <div className="checkout-payment-error__content">
            <span className="checkout-payment-error__title">Payment Attention</span>
            <p className="checkout-payment-error__msg">{paymentError}</p>
          </div>
          {createdPendingOrder && paymentMethod === 'online' && (
            <div className="checkout-payment-error__actions">
              {!createdPendingOrder.is_expired && (
                <Button
                  variant="outline"
                  size="sm"
                  leftIcon={<RefreshCw size={14} />}
                  onClick={() => executeRazorpayFlow(createdPendingOrder)}
                  data-testid="retry-payment-btn"
                  loading={isSubmitting}
                  disabled={isSubmitting}
                >
                  RETRY PAYMENT
                </Button>
              )}
              <Button
                variant={createdPendingOrder.is_expired ? 'primary' : 'ghost'}
                size="sm"
                onClick={handleCancelPendingOrder}
                data-testid="cancel-pending-btn"
                disabled={isSubmitting}
              >
                {createdPendingOrder.is_expired ? 'CANCEL EXPIRED ORDER' : 'CANCEL ORDER'}
              </Button>
            </div>
          )}
        </div>
      )}

      <div className="checkout-layout">
        {/* Main Steps Form */}
        <div className="checkout-steps">
          {/* STEP 1: Delivery Address */}
          <section className="checkout-section" aria-labelledby="step-address-heading">
            <div className="checkout-section__header">
              <div className="checkout-step-number">1</div>
              <h2 id="step-address-heading" className="checkout-section__title">
                Delivery Address
              </h2>
              <Button
                variant="outline"
                size="sm"
                leftIcon={<Plus size={14} />}
                onClick={() => setIsAddressModalOpen(true)}
                data-testid="add-address-btn"
              >
                ADD NEW
              </Button>
            </div>

            {addresses.length === 0 ? (
              <div className="checkout-address-empty">
                <MapPin size={28} />
                <p>No saved delivery addresses found.</p>
                <Button
                  variant="primary"
                  size="sm"
                  onClick={() => setIsAddressModalOpen(true)}
                >
                  ADD DELIVERY ADDRESS
                </Button>
              </div>
            ) : (
              <div className="checkout-address-list" role="radiogroup" aria-label="Delivery addresses">
                {addresses.map((addr) => {
                  const isSelected = addr.id === selectedAddressId;
                  return (
                    <label
                      key={addr.id}
                      className={`checkout-address-card ${isSelected ? 'is-selected' : ''}`}
                      data-testid={`address-card-${addr.id}`}
                    >
                      <input
                        type="radio"
                        name="delivery_address"
                        value={addr.id}
                        checked={isSelected}
                        onChange={() => setSelectedAddressId(addr.id)}
                        className="checkout-address-radio"
                      />
                      <div className="checkout-address-info">
                        <div className="checkout-address-name">
                          <strong>{addr.name}</strong>
                          {addr.is_default && (
                            <Badge variant="neutral" size="sm">
                              DEFAULT
                            </Badge>
                          )}
                        </div>
                        <p className="checkout-address-text">
                          {addr.address_line1}
                          {addr.address_line_2 ? `, ${addr.address_line_2}` : ''}
                        </p>
                        <p className="checkout-address-text">
                          {addr.city}{addr.state ? `, ${addr.state}` : ''}, {addr.postal_code}
                        </p>
                        <p className="checkout-address-phone">Phone: {addr.phone}</p>
                      </div>
                    </label>
                  );
                })}
              </div>
            )}
          </section>

          {/* STEP 2: Items Review */}
          <section className="checkout-section" aria-labelledby="step-review-heading">
            <div className="checkout-section__header">
              <div className="checkout-step-number">2</div>
              <h2 id="step-review-heading" className="checkout-section__title">
                Review Items ({displayItems.length})
              </h2>
            </div>

            <div className="checkout-items-preview">
              {displayItems.map((item) => (
                <div key={item.id || item.variant_id} className="checkout-item-row">
                  <img
                    src={item.image_url || '/placeholder-product.svg'}
                    alt={item.product_name}
                    className="checkout-item-thumb"
                  />
                  <div className="checkout-item-details">
                    <span className="checkout-item-name">{item.product_name}</span>
                    <span className="checkout-item-variant">
                      Size: {item.size || 'Standard'} • Color: {item.color || 'Standard'} • Qty:{' '}
                      {item.quantity}
                    </span>
                  </div>
                  <div className="checkout-item-price">
                    {formatCurrencyINR(
                      item.line_total || item.unit_price * item.quantity || item.subtotal || 0,
                    )}
                  </div>
                </div>
              ))}
            </div>
          </section>

          {/* STEP 3: Payment Method */}
          <section className="checkout-section" aria-labelledby="step-payment-heading">
            <div className="checkout-section__header">
              <div className="checkout-step-number">3</div>
              <h2 id="step-payment-heading" className="checkout-section__title">
                Payment Method
              </h2>
            </div>

            <div className="checkout-payment-options" role="radiogroup" aria-label="Payment method">
              {/* Online Razorpay */}
              <label
                className={`checkout-payment-option ${
                  paymentMethod === 'online' ? 'is-selected' : ''
                }`}
                data-testid="payment-method-online"
              >
                <input
                  type="radio"
                  name="payment_method"
                  value="online"
                  checked={paymentMethod === 'online'}
                  onChange={() => setPaymentMethod('online')}
                />
                <div className="checkout-payment-option__body">
                  <div className="checkout-payment-option__title">
                    <CreditCard size={18} />
                    <strong>Online Payment</strong>
                    <Badge variant="primary" size="sm">
                      RECOMMENDED
                    </Badge>
                  </div>
                  <p className="checkout-payment-option__desc">
                    UPI, Credit/Debit Cards, NetBanking via secure Razorpay checkout.
                  </p>
                </div>
              </label>

              {/* Cash On Delivery */}
              <label
                className={`checkout-payment-option ${
                  paymentMethod === 'cod' ? 'is-selected' : ''
                } ${createdPendingOrder ? 'is-disabled' : ''}`}
                data-testid="payment-method-cod"
              >
                <input
                  type="radio"
                  name="payment_method"
                  value="cod"
                  disabled={Boolean(createdPendingOrder)}
                  checked={paymentMethod === 'cod'}
                  onChange={() => setPaymentMethod('cod')}
                />
                <div className="checkout-payment-option__body">
                  <div className="checkout-payment-option__title">
                    <Truck size={18} />
                    <strong>Cash on Delivery (COD)</strong>
                  </div>
                  <p className="checkout-payment-option__desc">
                    {createdPendingOrder
                      ? 'Unavailable while a pending online order is awaiting payment or cancellation.'
                      : 'Pay in cash upon doorstep delivery. Subject to courier verification.'}
                  </p>
                </div>
              </label>

              {/* Wallet */}
              <label
                className={`checkout-payment-option ${
                  paymentMethod === 'wallet' ? 'is-selected' : ''
                } ${!canUseWallet || createdPendingOrder ? 'is-disabled' : ''}`}
                data-testid="payment-method-wallet"
              >
                <input
                  type="radio"
                  name="payment_method"
                  value="wallet"
                  disabled={!canUseWallet || Boolean(createdPendingOrder)}
                  checked={paymentMethod === 'wallet'}
                  onChange={() => setPaymentMethod('wallet')}
                />
                <div className="checkout-payment-option__body">
                  <div className="checkout-payment-option__title">
                    <Wallet size={18} />
                    <strong>Drape Atelier Wallet</strong>
                    <span className="checkout-wallet-balance">
                      Balance: {wallet ? formatCurrencyINR(wallet.balance) : '₹0.00'}
                    </span>
                  </div>
                  <p className="checkout-payment-option__desc">
                    {createdPendingOrder
                      ? 'Unavailable while a pending online order is awaiting payment or cancellation.'
                      : !canUseWallet
                      ? 'Insufficient wallet balance for this order total.'
                      : 'Instant one-click checkout using your Drape account balance.'}
                  </p>
                </div>
              </label>
            </div>
          </section>
        </div>

        {/* Sidebar Summary & Order Action */}
        <aside className="checkout-summary-sidebar" aria-label="Order summary">
          <div className="checkout-summary-card">
            <h2 className="checkout-summary-title">Order Summary</h2>

            {/* Coupon Application Box */}
            <div className="checkout-coupon-box">
              {appliedCoupon ? (
                <div className="checkout-applied-coupon">
                  <div className="checkout-applied-coupon__info">
                    <Tag size={16} />
                    <span>
                      Coupon <strong>{appliedCoupon.code}</strong> applied
                    </span>
                  </div>
                  <button
                    type="button"
                    className="checkout-remove-coupon-btn"
                    onClick={handleRemoveCoupon}
                  >
                    Remove
                  </button>
                </div>
              ) : (
                <form onSubmit={handleApplyCoupon} className="checkout-coupon-form">
                  <Input
                    placeholder="Enter Coupon / Promo Code"
                    value={couponCode}
                    onChange={(e) => setCouponCode(e.target.value)}
                    size="sm"
                    data-testid="coupon-input"
                  />
                  <Button
                    type="submit"
                    variant="outline"
                    size="sm"
                    loading={isApplyingCoupon}
                    disabled={isApplyingCoupon || !couponCode.trim()}
                    data-testid="apply-coupon-btn"
                  >
                    APPLY
                  </Button>
                </form>
              )}
              {couponError && <p className="checkout-coupon-error">{couponError}</p>}
            </div>

            {/* Summary Line Items */}
            <div className="checkout-summary-rows">
              <div className="checkout-summary-row">
                <span>Subtotal</span>
                <span>{formatCurrencyINR(rawSubtotal)}</span>
              </div>

              {discountAmount > 0 && (
                <div className="checkout-summary-row checkout-summary-discount">
                  <span>Promotion Discount</span>
                  <span>-{formatCurrencyINR(discountAmount)}</span>
                </div>
              )}

              <div className="checkout-summary-row">
                <span>Shipping</span>
                <span className="checkout-free-badge">COMPLIMENTARY</span>
              </div>

              <div className="checkout-summary-divider" />

              <div className="checkout-summary-row checkout-summary-total">
                <span>Total Amount</span>
                <span>{formatCurrencyINR(estimatedTotal)}</span>
              </div>
            </div>

            <Button
              variant="primary"
              size="lg"
              fullWidth
              loading={isSubmitting}
              disabled={
                isSubmitting ||
                Boolean(createdPendingOrder?.is_expired) ||
                (!selectedAddressId && !createdPendingOrder) ||
                (displayItems.length === 0 && !createdPendingOrder)
              }
              onClick={handleSubmitOrder}
              data-testid="place-order-btn"
            >
              {isSubmitting
                ? 'PROCESSING...'
                : createdPendingOrder?.is_expired
                ? 'PAYMENT EXPIRED — PLEASE CANCEL'
                : createdPendingOrder
                ? `COMPLETE PAYMENT (${formatCurrencyINR(estimatedTotal)})`
                : `PLACE ORDER (${formatCurrencyINR(estimatedTotal)})`}
            </Button>

            <div className="checkout-security-guarantee">
              <ShieldCheck size={18} />
              <span>Complimentary insured shipping & encrypted transactions</span>
            </div>
          </div>
        </aside>
      </div>

      {/* Add New Address Modal */}
      <Modal
        isOpen={isAddressModalOpen}
        onClose={() => setIsAddressModalOpen(false)}
        title="Add Delivery Address"
      >
        <form onSubmit={handleSaveAddress} className="checkout-address-modal-form">
          {addressFormError && (
            <div className="checkout-modal-error">{addressFormError}</div>
          )}

          <div className="checkout-modal-form-grid">
            <FormField label="Recipient Full Name" htmlFor="addr-name" required>
              <Input
                id="addr-name"
                value={newAddress.name}
                onChange={(e) => setNewAddress({ ...newAddress, name: e.target.value })}
                placeholder="e.g. Eleanor Vance"
                required
              />
            </FormField>

            <FormField label="Phone Number" htmlFor="addr-phone" required>
              <Input
                id="addr-phone"
                type="tel"
                value={newAddress.phone}
                onChange={(e) => setNewAddress({ ...newAddress, phone: e.target.value })}
                placeholder="e.g. +91 9876543210"
                required
              />
            </FormField>
          </div>

          <FormField label="Address Line 1" htmlFor="addr-line1" required>
            <Input
              id="addr-line1"
              value={newAddress.address_line1}
              onChange={(e) => setNewAddress({ ...newAddress, address_line1: e.target.value })}
              placeholder="e.g. Flat 4B, Kensington Heights"
              required
            />
          </FormField>

          <FormField label="Address Line 2 (Optional)" htmlFor="addr-line2">
            <Input
              id="addr-line2"
              value={newAddress.address_line_2}
              onChange={(e) => setNewAddress({ ...newAddress, address_line_2: e.target.value })}
              placeholder="e.g. Near Central Avenue"
            />
          </FormField>

          <div className="checkout-modal-form-grid">
            <FormField label="City" htmlFor="addr-city" required>
              <Input
                id="addr-city"
                value={newAddress.city}
                onChange={(e) => setNewAddress({ ...newAddress, city: e.target.value })}
                placeholder="e.g. Mumbai"
                required
              />
            </FormField>

            <FormField label="State / Province" htmlFor="addr-state" required>
              <Input
                id="addr-state"
                value={newAddress.state}
                onChange={(e) => setNewAddress({ ...newAddress, state: e.target.value })}
                placeholder="e.g. Maharashtra"
                required
              />
            </FormField>
          </div>

          <FormField label="Postal Code" htmlFor="addr-postal" required>
            <Input
              id="addr-postal"
              value={newAddress.postal_code}
              onChange={(e) => setNewAddress({ ...newAddress, postal_code: e.target.value })}
              placeholder="e.g. 400001"
              required
            />
          </FormField>

          <label className="checkout-modal-checkbox">
            <input
              type="checkbox"
              checked={newAddress.is_default}
              onChange={(e) => setNewAddress({ ...newAddress, is_default: e.target.checked })}
            />
            <span>Set as default delivery address</span>
          </label>

          <div className="checkout-modal-actions">
            <Button
              type="button"
              variant="outline"
              size="md"
              onClick={() => setIsAddressModalOpen(false)}
            >
              CANCEL
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="md"
              loading={isSavingAddress}
              disabled={isSavingAddress}
            >
              SAVE ADDRESS
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
