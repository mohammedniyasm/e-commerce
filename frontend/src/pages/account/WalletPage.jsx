import React, { useState, useEffect, useCallback } from 'react';
import {
  Wallet,
  ArrowUpRight,
  ArrowDownLeft,
  Plus,
  RefreshCw,
  Clock,
  CheckCircle2,
  AlertCircle,
  Shield,
  CreditCard,
  ChevronLeft,
  ChevronRight,
} from 'lucide-react';
import {
  Button,
  StatusPill,
  Badge,
  SkeletonLoader,
  Modal,
  FormField,
  Input,
} from '../../components/common';
import { useToast } from '../../context/ToastContext';
import { useAuth } from '../../context/AuthContext';
import {
  getWallet,
  getWalletTransactions,
  createTopUpOrder,
  verifyTopUpPayment,
} from '../../api/walletService';
import { formatCurrencyINR, formatDate } from '../../utils/formatters';
import './WalletPage.css';

const PRESET_AMOUNTS = [1000, 2500, 5000, 10000];

export function WalletPage() {
  const { user } = useAuth();
  const { showToast } = useToast();

  const [wallet, setWallet] = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [totalTransactions, setTotalTransactions] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [isLoading, setIsLoading] = useState(true);

  // Top-Up Modal State
  const [isTopUpModalOpen, setIsTopUpModalOpen] = useState(false);
  const [topUpAmount, setTopUpAmount] = useState(2500);
  const [customAmount, setCustomAmount] = useState('');
  const [isTopUpLoading, setIsTopUpLoading] = useState(false);
  const [topUpError, setTopUpError] = useState('');

  const loadWalletData = useCallback(
    async (page = 1) => {
      setIsLoading(true);
      try {
        const [walletRes, txRes] = await Promise.allSettled([
          getWallet(),
          getWalletTransactions({ page, limit: 10 }),
        ]);

        if (walletRes.status === 'fulfilled' && walletRes.value) {
          setWallet(walletRes.value);
        }

        if (txRes.status === 'fulfilled' && txRes.value) {
          const txData = txRes.value;
          setTransactions(txData.transactions || []);
          setTotalTransactions(txData.total || 0);
          setCurrentPage(txData.page || page);
        }
      } catch (err) {
        showToast({
          title: 'Wallet Unavailable',
          message: err.message || 'Unable to retrieve wallet data.',
          type: 'danger',
        });
      } finally {
        setIsLoading(false);
      }
    },
    [showToast],
  );

  useEffect(() => {
    loadWalletData(currentPage);
  }, [loadWalletData, currentPage]);

  // Razorpay Top-Up Flow
  async function handleTopUpSubmit(e) {
    e.preventDefault();
    const finalAmount = customAmount ? Number(customAmount) : Number(topUpAmount);

    if (isNaN(finalAmount) || finalAmount < 100) {
      setTopUpError('Minimum top-up amount is ₹100.');
      return;
    }
    if (finalAmount > 100000) {
      setTopUpError('Maximum single top-up amount is ₹1,00,000.');
      return;
    }

    setTopUpError('');
    setIsTopUpLoading(true);

    try {
      const orderRes = await createTopUpOrder(finalAmount);
      const { razorpay_key_id, gateway_order_id, amount, currency } = orderRes;

      if (!window.Razorpay) {
        // Dynamically load Razorpay SDK if not already in window
        await new Promise((resolve, reject) => {
          const script = document.createElement('script');
          script.src = 'https://checkout.razorpay.com/v1/checkout.js';
          script.onload = resolve;
          script.onerror = () => reject(new Error('Failed to load Razorpay payment gateway.'));
          document.body.appendChild(script);
        });
      }

      const options = {
        key: razorpay_key_id || 'rzp_test_key',
        amount: amount,
        currency: currency || 'INR',
        name: 'DRAPE',
        description: 'Wallet top-up',
        order_id: gateway_order_id,
        prefill: {
          name: user?.name || '',
          email: user?.email || '',
          contact: user?.phone || '',
        },
        theme: {
          color: '#0F172A',
        },
        modal: {
          ondismiss: () => {
            setIsTopUpLoading(false);
            showToast({
              title: 'Top-Up Dismissed',
              message: 'Payment window was closed.',
              type: 'info',
            });
          },
        },
        handler: async (response) => {
          try {
            await verifyTopUpPayment({
              gateway_order_id,
              gateway_payment_id: response.razorpay_payment_id,
              razorpay_signature: response.razorpay_signature,
            });

            setIsTopUpModalOpen(false);
            setCustomAmount('');
            showToast({
              title: 'Wallet Funded',
              message: `${formatCurrencyINR(finalAmount)} added to your wallet.`,
              type: 'success',
            });
            await loadWalletData(1);
          } catch (verifyErr) {
            setTopUpError(verifyErr.message || 'Payment verification failed.');
          } finally {
            setIsTopUpLoading(false);
          }
        },
      };

      const rzp = new window.Razorpay(options);
      rzp.on('payment.failed', (resp) => {
        setTopUpError(resp.error?.description || 'Payment failed.');
        setIsTopUpLoading(false);
      });
      rzp.open();
    } catch (err) {
      setTopUpError(err.message || 'Unable to initiate top-up.');
      setIsTopUpLoading(false);
    }
  }

  const balance = wallet?.balance ?? 0;
  const totalPages = Math.ceil(totalTransactions / 10) || 1;

  return (
    <div className="account-wallet-page" data-testid="account-wallet-page">
      {/* Top Header */}
      <div className="account-wallet-header">
        <div>
          <h1 className="account-wallet-title">Wallet</h1>
          <p className="account-wallet-subtitle">
            Use your balance for orders and receive eligible refunds and referral rewards here.
          </p>
        </div>
        <Button
          variant="primary"
          size="sm"
          leftIcon={<Plus size={16} />}
          onClick={() => {
            setTopUpError('');
            setIsTopUpModalOpen(true);
          }}
          data-testid="top-up-wallet-btn"
        >
          TOP UP WALLET
        </Button>
      </div>

      {/* Hero Wallet Card */}
      <div className="wallet-hero-card" data-testid="wallet-hero-card">
        <div className="wallet-hero-left">
          <div className="wallet-hero-badge">
            <Shield size={14} />
            <span>SECURE WALLET</span>
          </div>
          <span className="wallet-hero-label">AVAILABLE BALANCE</span>
          <h2 className="wallet-hero-balance" data-testid="wallet-balance">
            {isLoading ? '₹ —' : formatCurrencyINR(balance)}
          </h2>
          <p className="wallet-hero-note">
            Your wallet balance can be used at checkout.
          </p>
        </div>

        <div className="wallet-hero-actions">
          <Button
            variant="outline"
            size="md"
            onClick={() => loadWalletData(currentPage)}
            leftIcon={<RefreshCw size={14} />}
            data-testid="refresh-wallet-btn"
          >
            REFRESH
          </Button>
        </div>
      </div>

      {/* Transaction History Section */}
      <section className="account-card" aria-labelledby="tx-heading">
        <div className="account-card__header">
          <div className="account-card__title-group">
            <Wallet size={18} className="account-card__icon" aria-hidden="true" />
            <h2 id="tx-heading" className="account-card__title">
              Transaction History
            </h2>
          </div>
        </div>

        <div className="account-card__body">
          {isLoading ? (
            <SkeletonLoader lines={6} />
          ) : transactions.length > 0 ? (
            <div className="wallet-tx-table" data-testid="wallet-tx-table">
              {transactions.map((tx) => {
                const isCredit = (tx.type || tx.transaction_type || '').toLowerCase() === 'credit';
                const status = (tx.status || 'completed').toLowerCase();
                const source = (tx.source || tx.source_type || 'adjustment').replace(/_/g, ' ');

                return (
                  <div key={tx.id} className="wallet-tx-row" data-testid={`wallet-tx-${tx.id}`}>
                    <div className="wallet-tx-icon-wrap" aria-hidden="true">
                      {isCredit ? (
                        <div className="wallet-tx-icon is-credit">
                          <ArrowDownLeft size={16} />
                        </div>
                      ) : (
                        <div className="wallet-tx-icon is-debit">
                          <ArrowUpRight size={16} />
                        </div>
                      )}
                    </div>

                    <div className="wallet-tx-info">
                      <div className="wallet-tx-title-row">
                        <strong className="wallet-tx-title">
                          {tx.description || source.toUpperCase()}
                        </strong>
                        <StatusPill
                          status={status}
                          variant={status === 'completed' ? 'success' : 'warning'}
                        />
                      </div>
                      <div className="wallet-tx-meta">
                        <span>Transaction #{tx.id}</span>
                        <span>•</span>
                        <span>{formatDate(tx.created_at)}</span>
                        <span>•</span>
                        <span className="wallet-tx-source">{source}</span>
                      </div>
                    </div>

                    <div className="wallet-tx-amount-col">
                      <strong className={`wallet-tx-amount ${isCredit ? 'is-credit' : 'is-debit'}`}>
                        {isCredit ? '+ ' : '- '}
                        {formatCurrencyINR(tx.amount)}
                      </strong>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="account-empty-block" data-testid="wallet-empty-state">
              <Clock size={36} className="account-empty-icon" />
              <h2 className="account-empty-title">No Transactions Yet</h2>
              <p className="account-empty-text">
                Your wallet activity will appear here, including top-ups, payments, refunds, and referral rewards.
              </p>
            </div>
          )}
        </div>
      </section>

      {/* Pagination Controls */}
      {totalPages > 1 && (
        <div className="orders-pagination" data-testid="wallet-pagination">
          <Button
            variant="outline"
            size="sm"
            disabled={currentPage <= 1}
            onClick={() => setCurrentPage((p) => Math.max(1, p - 1))}
            leftIcon={<ChevronLeft size={14} />}
          >
            PREVIOUS
          </Button>
          <span className="orders-pagination-info">
            Page {currentPage} of {totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            disabled={currentPage >= totalPages}
            onClick={() => setCurrentPage((p) => Math.min(totalPages, p + 1))}
            rightIcon={<ChevronRight size={14} />}
          >
            NEXT
          </Button>
        </div>
      )}

      {/* Top Up Modal */}
      <Modal
        isOpen={isTopUpModalOpen}
        onClose={() => setIsTopUpModalOpen(false)}
        title="Top Up Wallet"
        eyebrow="ADD FUNDS"
        size="md"
      >
        <form onSubmit={handleTopUpSubmit} className="wallet-topup-form" noValidate>
          <p className="wallet-topup-desc">
            Choose an amount to add to your wallet using Razorpay.
          </p>

          {topUpError && (
            <div className="wallet-topup-error" role="alert">
              <AlertCircle size={14} />
              <span>{topUpError}</span>
            </div>
          )}

          <div className="wallet-preset-grid">
            {PRESET_AMOUNTS.map((amt) => {
              const isSelected = !customAmount && topUpAmount === amt;
              return (
                <button
                  key={amt}
                  type="button"
                  className={`wallet-preset-btn ${isSelected ? 'is-selected' : ''}`}
                  onClick={() => {
                    setTopUpAmount(amt);
                    setCustomAmount('');
                  }}
                  data-testid={`preset-amt-${amt}`}
                >
                  {formatCurrencyINR(amt)}
                </button>
              );
            })}
          </div>

          <FormField label="OR ENTER CUSTOM AMOUNT (INR ₹)" htmlFor="custom-topup-amount">
            <Input
              id="custom-topup-amount"
              type="number"
              min={100}
              max={100000}
              value={customAmount}
              onChange={(e) => {
                setCustomAmount(e.target.value);
                setTopUpAmount(0);
              }}
              placeholder="e.g. 7500"
              data-testid="custom-topup-input"
            />
          </FormField>

          <div className="wallet-topup-summary">
            <span>Funding Amount:</span>
            <strong className="wallet-topup-final-amt">
              {formatCurrencyINR(customAmount ? Number(customAmount) : topUpAmount)}
            </strong>
          </div>

          <div className="wallet-topup-actions">
            <Button
              type="button"
              variant="outline"
              size="md"
              onClick={() => setIsTopUpModalOpen(false)}
              disabled={isTopUpLoading}
            >
              CANCEL
            </Button>
            <Button
              type="submit"
              variant="primary"
              size="md"
              isLoading={isTopUpLoading}
              data-testid="proceed-topup-btn"
            >
              PROCEED TO RAZORPAY
            </Button>
          </div>
        </form>
      </Modal>
    </div>
  );
}
