import { useEffect, useState, useCallback } from 'react';
import {
  Users,
  Search,
  UserPlus,
  Lock,
  Unlock,
  Wallet,
  Mail,
  Phone,
  RefreshCw,
  Eye,
  ShieldAlert,
  Calendar,
} from 'lucide-react';
import {
  listAdminUsers,
  addAdminCustomer,
  blockAdminUser,
  unblockAdminUser,
  getUserWallet,
} from '../../../api/adminOperationsService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { Modal } from '../../../components/common/Modal';
import { useToast } from '../../../context/ToastContext';
import { formatCurrency, formatDate } from '../../../utils/formatters';
import './AdminCustomersPage.css';

export function AdminCustomersPage() {
  const { showToast } = useToast();

  const [customers, setCustomers] = useState([]);
  const [isLoading, setIsLoading] = useState(true);

  // Pagination & Search
  const [page, setPage] = useState(1);
  const [limit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');

  // Add Customer Modal
  const [isAddModalOpen, setIsAddModalOpen] = useState(false);
  const [newName, setNewName] = useState('');
  const [newEmail, setNewEmail] = useState('');
  const [newPhone, setNewPhone] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [isAddingCustomer, setIsAddingCustomer] = useState(false);

  // Block / Unblock Modal
  const [userToToggleBlock, setUserToToggleBlock] = useState(null);
  const [isTogglingBlock, setIsTogglingBlock] = useState(false);

  // Customer Wallet Modal
  const [viewingCustomerWallet, setViewingCustomerWallet] = useState(null);
  const [walletData, setWalletData] = useState(null);
  const [isLoadingWallet, setIsLoadingWallet] = useState(false);

  const fetchCustomers = useCallback(async () => {
    setIsLoading(true);
    try {
      const data = await listAdminUsers({
        page,
        limit,
        search,
      });

      setCustomers(data?.users ?? []);
      setTotalItems(data?.total ?? 0);
      setTotalPages(data?.total_pages ?? (Math.ceil((data?.total ?? 0) / limit) || 1));
    } catch (err) {
      console.error('Failed to load customers:', err);
      showToast(err?.response?.data?.message || 'Failed to load customers', 'error');
    } finally {
      setIsLoading(false);
    }
  }, [page, limit, search, showToast]);

  useEffect(() => {
    fetchCustomers();
  }, [fetchCustomers]);

  const handleSearchSubmit = (e) => {
    e.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  };

  const handleAddCustomerSubmit = async (e) => {
    e.preventDefault();
    if (!newName.trim() || !newEmail.trim() || !newPassword.trim()) {
      showToast('Name, email, and initial password are required', 'error');
      return;
    }

    setIsAddingCustomer(true);
    try {
      await addAdminCustomer({
        name: newName.trim(),
        email: newEmail.trim().toLowerCase(),
        phone: newPhone.trim() || undefined,
        password: newPassword.trim(),
      });
      showToast('Customer account created successfully', 'success');
      setIsAddModalOpen(false);
      setNewName('');
      setNewEmail('');
      setNewPhone('');
      setNewPassword('');
      await fetchCustomers();
    } catch (err) {
      console.error('Failed to add customer:', err);
      showToast(err?.response?.data?.message || 'Failed to create customer', 'error');
    } finally {
      setIsAddingCustomer(false);
    }
  };

  const handleConfirmToggleBlock = async () => {
    if (!userToToggleBlock) return;
    const isCurrentlyBlocked = userToToggleBlock.is_blocked;
    setIsTogglingBlock(true);
    try {
      if (isCurrentlyBlocked) {
        await unblockAdminUser(userToToggleBlock.id);
        showToast(`Customer ${userToToggleBlock.name} unblocked`, 'success');
      } else {
        await blockAdminUser(userToToggleBlock.id);
        showToast(`Customer ${userToToggleBlock.name} blocked`, 'success');
      }
      setUserToToggleBlock(null);
      await fetchCustomers();
    } catch (err) {
      console.error('Failed to update customer block status:', err);
      showToast(err?.response?.data?.message || 'Failed to update access', 'error');
    } finally {
      setIsTogglingBlock(false);
    }
  };

  const handleOpenWalletModal = async (customer) => {
    setViewingCustomerWallet(customer);
    setIsLoadingWallet(true);
    setWalletData(null);
    try {
      const data = await getUserWallet(customer.id);
      setWalletData(data);
    } catch (err) {
      console.error('Failed to load user wallet:', err);
      showToast('No wallet balance found or error loading wallet', 'info');
    } finally {
      setIsLoadingWallet(false);
    }
  };

  const columns = [
    {
      key: 'name',
      header: 'Customer',
      render: (val, user) => {
        const u = (typeof val === 'object' && val !== null) ? val : (user || {});
        return (
          <div className="drape-customer-cell">
            <span className="font-semibold">{u.name || (typeof val === 'string' ? val : 'Customer')}</span>
            <span className="drape-customer-email">{u.email || '—'}</span>
          </div>
        );
      },
    },
    {
      key: 'phone',
      header: 'Phone',
      render: (val, user) => {
        const u = (typeof val === 'object' && val !== null) ? val : (user || {});
        const phone = typeof val === 'string' ? val : u.phone;
        return (
          <span className="drape-phone-text">{phone || '—'}</span>
        );
      },
    },
    {
      key: 'verified',
      header: 'Email Verified',
      align: 'center',
      render: (val, user) => {
        const u = (typeof val === 'object' && val !== null) ? val : (user || {});
        const isVerified =
          typeof val === 'boolean'
            ? val
            : Boolean(u.is_email_verified ?? (u.email_verified_at != null));
        return (
          <Badge variant={isVerified ? 'success' : 'neutral'}>
            {isVerified ? 'Verified' : 'Unverified'}
          </Badge>
        );
      },
    },
    {
      key: 'status',
      header: 'Access',
      align: 'center',
      render: (val, user) => {
        const u = (typeof val === 'object' && val !== null) ? val : (user || {});
        const isBlocked = Boolean(u.is_blocked);
        return (
          <Badge variant={isBlocked ? 'danger' : 'success'}>
            {isBlocked ? 'Blocked' : 'Active'}
          </Badge>
        );
      },
    },
    {
      key: 'registered',
      header: 'Joined',
      render: (val, user) => {
        const u = (typeof val === 'object' && val !== null) ? val : (user || {});
        const date = typeof val === 'string' ? val : u.created_at;
        return (
          <span className="drape-joined-date">{formatDate(date)}</span>
        );
      },
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (val, user) => {
        const u = (typeof val === 'object' && val !== null) ? val : (user || {});
        return (
          <div className="drape-table-actions">
            <Button
              variant="ghost"
              size="sm"
              icon={<Wallet size={13} />}
              onClick={() => handleOpenWalletModal(u)}
              title="View Wallet Balance & Transactions"
            >
              Wallet
            </Button>

            <Button
              variant="ghost"
              size="sm"
              icon={u.is_blocked ? <Unlock size={13} /> : <Lock size={13} />}
              onClick={() => setUserToToggleBlock(u)}
              title={u.is_blocked ? 'Unblock customer' : 'Block customer'}
            >
              {u.is_blocked ? 'Unblock' : 'Block'}
            </Button>
          </div>
        );
      },
    },
  ];

  return (
    <div className="admin-customers-page">
      {/* Header */}
      <div className="admin-page-header">
        <div>
          <span className="admin-page-eyebrow">Directory & Access</span>
          <h1 className="admin-page-title">Customer Management</h1>
          <p className="admin-page-description">
            View registered customers, manage account security, and audit individual wallet balances.
          </p>
        </div>
        <div className="admin-header-actions">
          <Button
            variant="outline"
            size="sm"
            icon={<RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />}
            onClick={fetchCustomers}
            disabled={isLoading}
          >
            Refresh
          </Button>
          <Button
            variant="primary"
            size="sm"
            icon={<UserPlus size={14} />}
            onClick={() => setIsAddModalOpen(true)}
          >
            Add Customer
          </Button>
        </div>
      </div>

      {/* Toolbar */}
      <div className="admin-toolbar">
        <form onSubmit={handleSearchSubmit} className="admin-search-form">
          <div className="admin-search-input-wrapper">
            <Search size={16} className="admin-search-icon" />
            <input
              type="text"
              placeholder="Search by customer name, email, or phone..."
              value={searchInput}
              onChange={(e) => setSearchInput(e.target.value)}
              className="admin-search-input"
            />
          </div>
          <Button type="submit" variant="secondary" size="sm">
            Search
          </Button>
        </form>

        {search && (
          <Button
            variant="ghost"
            size="sm"
            onClick={() => {
              setSearch('');
              setSearchInput('');
              setPage(1);
            }}
          >
            Clear Search
          </Button>
        )}
      </div>

      {/* Table */}
      <DataTable
        columns={columns}
        data={customers}
        keyField="id"
        isLoading={isLoading}
        emptyMessage="No customer accounts found."
        pagination={{
          page,
          totalPages,
          totalItems,
          limit,
          onPageChange: (newPage) => setPage(newPage),
        }}
      />

      {/* Add Customer Modal */}
      <Modal
        isOpen={isAddModalOpen}
        onClose={() => setIsAddModalOpen(false)}
        title="Add Customer Account"
      >
        <form onSubmit={handleAddCustomerSubmit} className="admin-modal-form">
          <label className="admin-form-label">
            Full Name *
            <input
              type="text"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              placeholder="e.g. Eleanor Vance"
              required
              className="admin-input-full"
            />
          </label>

          <label className="admin-form-label">
            Email Address *
            <input
              type="email"
              value={newEmail}
              onChange={(e) => setNewEmail(e.target.value)}
              placeholder="customer@example.com"
              required
              className="admin-input-full"
            />
          </label>

          <label className="admin-form-label">
            Phone Number
            <input
              type="tel"
              value={newPhone}
              onChange={(e) => setNewPhone(e.target.value)}
              placeholder="+91 9876543210"
              className="admin-input-full"
            />
          </label>

          <label className="admin-form-label">
            Initial Password *
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              placeholder="Secure password (min 6 characters)"
              required
              minLength={6}
              className="admin-input-full"
            />
          </label>

          <div className="admin-modal-actions">
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsAddModalOpen(false)}
            >
              Cancel
            </Button>
            <Button type="submit" variant="primary" disabled={isAddingCustomer}>
              {isAddingCustomer ? 'Creating...' : 'Create Account'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Block / Unblock Modal */}
      <Modal
        isOpen={Boolean(userToToggleBlock)}
        onClose={() => setUserToToggleBlock(null)}
        title={userToToggleBlock?.is_blocked ? 'Unblock Customer Account' : 'Block Customer Account'}
      >
        {userToToggleBlock && (
          <div className="admin-modal-content">
            <p>
              Are you sure you want to {userToToggleBlock.is_blocked ? 'unblock' : 'block'}{' '}
              <strong>{userToToggleBlock.name}</strong> ({userToToggleBlock.email})?
            </p>
            <p className="text-muted">
              {userToToggleBlock.is_blocked
                ? 'The customer will immediately regain access to login and checkout.'
                : 'The customer will be immediately barred from placing orders and logging in.'}
            </p>
            <div className="admin-modal-actions">
              <Button
                type="button"
                variant="outline"
                onClick={() => setUserToToggleBlock(null)}
              >
                Cancel
              </Button>
              <Button
                type="button"
                variant={userToToggleBlock.is_blocked ? 'primary' : 'danger'}
                onClick={handleConfirmToggleBlock}
                disabled={isTogglingBlock}
              >
                {isTogglingBlock
                  ? 'Processing...'
                  : userToToggleBlock.is_blocked
                  ? 'Confirm Unblock'
                  : 'Confirm Block'}
              </Button>
            </div>
          </div>
        )}
      </Modal>

      {/* Customer Wallet Modal */}
      <Modal
        isOpen={Boolean(viewingCustomerWallet)}
        onClose={() => setViewingCustomerWallet(null)}
        title={`Customer Wallet: ${viewingCustomerWallet?.name}`}
      >
        <div className="admin-modal-content">
          {isLoadingWallet ? (
            <p>Loading wallet balance...</p>
          ) : walletData ? (
            <div className="admin-wallet-info">
              <div className="admin-wallet-balance-card">
                <span className="admin-wallet-label">Current Balance</span>
                <span className="admin-wallet-amount">
                  {formatCurrency(walletData.balance || 0)}
                </span>
              </div>

              {walletData.transactions && walletData.transactions.length > 0 ? (
                <div className="admin-wallet-txns">
                  <h4>Recent Wallet Activity</h4>
                  <div className="admin-usages-table-wrapper">
                    <table className="admin-items-table">
                      <thead>
                        <tr>
                          <th>Type</th>
                          <th className="text-right">Amount</th>
                          <th>Description</th>
                          <th>Date</th>
                        </tr>
                      </thead>
                      <tbody>
                        {walletData.transactions.map((tx, idx) => (
                          <tr key={tx.id || idx}>
                            <td>
                              <Badge
                                variant={
                                  tx.type === 'credit' || tx.type === 'refund'
                                    ? 'success'
                                    : 'neutral'
                                }
                              >
                                {tx.type}
                              </Badge>
                            </td>
                            <td className="text-right font-semibold">
                              {tx.type === 'credit' || tx.type === 'refund' ? '+' : '-'}
                              {formatCurrency(tx.amount)}
                            </td>
                            <td>{tx.description || 'Wallet transaction'}</td>
                            <td>{formatDate(tx.created_at)}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                </div>
              ) : (
                <p className="text-muted">No wallet transactions recorded yet.</p>
              )}
            </div>
          ) : (
            <p className="text-muted">No wallet details found for this customer.</p>
          )}

          <div className="admin-modal-actions">
            <Button
              type="button"
              variant="outline"
              onClick={() => setViewingCustomerWallet(null)}
            >
              Close
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
