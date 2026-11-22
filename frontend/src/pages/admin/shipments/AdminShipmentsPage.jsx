import { useCallback, useEffect, useRef, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { Clock3, Eye, Filter, Package, RefreshCw, Search, Truck } from 'lucide-react';
import {
  createShipment,
  getShipmentById,
  getShipmentByOrderId,
  getShipmentTrackingHistory,
  listAdminShipments,
  updateShipment,
  updateShippingStatus,
} from '../../../api/adminOperationsService';
import { Badge } from '../../../components/common/Badge';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { Modal } from '../../../components/common/Modal';
import { useToast } from '../../../context/ToastContext';
import { formatDate } from '../../../utils/formatters';
import {
  getShipmentStatusTransitions,
  SHIPMENT_STATUS_LABELS,
  SHIPMENT_TYPE_LABELS,
} from '../../../utils/shipmentStatus';
import './AdminShipmentsPage.css';

const ORDER_ID_PATTERN = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

function getHistoryItems(historyResponse, shipmentHistory) {
  if (Array.isArray(historyResponse)) return historyResponse;
  if (Array.isArray(historyResponse?.history)) return historyResponse.history;
  return Array.isArray(shipmentHistory) ? shipmentHistory : [];
}

export function AdminShipmentsPage() {
  const [searchParams] = useSearchParams();
  const { showToast } = useToast();
  const initialOrderId = searchParams.get('orderId') || '';

  const [shipments, setShipments] = useState([]);
  const [isListLoading, setIsListLoading] = useState(true);
  const [listError, setListError] = useState('');
  const [isDetailsOpen, setIsDetailsOpen] = useState(false);
  const [page, setPage] = useState(1);
  const [limit, setLimit] = useState(10);
  const [totalPages, setTotalPages] = useState(1);
  const [totalItems, setTotalItems] = useState(0);
  const [search, setSearch] = useState('');
  const [searchInput, setSearchInput] = useState('');
  const [selectedStatus, setSelectedStatus] = useState('');
  const [selectedType, setSelectedType] = useState('');
  const [sort, setSort] = useState('newest');
  const [advancedFilters, setAdvancedFilters] = useState({
    courier_name: '',
    order_id: initialOrderId,
    return_request_id: '',
    created_from: '',
    created_to: '',
  });
  const [appliedAdvancedFilters, setAppliedAdvancedFilters] = useState({
    courier_name: '',
    order_id: initialOrderId,
    return_request_id: '',
    created_from: '',
    created_to: '',
  });

  const [lookupType, setLookupType] = useState(initialOrderId ? 'order' : 'shipment');
  const [lookupValue, setLookupValue] = useState(initialOrderId);
  const [shipment, setShipment] = useState(null);
  const [history, setHistory] = useState([]);
  const [hasSearched, setHasSearched] = useState(false);
  const [notFoundLookup, setNotFoundLookup] = useState(null);
  const [isLoading, setIsLoading] = useState(false);

  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [createOrderId, setCreateOrderId] = useState(initialOrderId);
  const [courierName, setCourierName] = useState('');
  const [trackingNumber, setTrackingNumber] = useState('');
  const [isCreating, setIsCreating] = useState(false);

  const [isEditOpen, setIsEditOpen] = useState(false);
  const [editCourierName, setEditCourierName] = useState('');
  const [editTrackingNumber, setEditTrackingNumber] = useState('');
  const [isSavingTracking, setIsSavingTracking] = useState(false);

  const [isStatusOpen, setIsStatusOpen] = useState(false);
  const [targetStatus, setTargetStatus] = useState('');
  const [statusLocation, setStatusLocation] = useState('');
  const [statusDescription, setStatusDescription] = useState('');
  const [isUpdatingStatus, setIsUpdatingStatus] = useState(false);
  const shipmentListRequestId = useRef(0);

  const fetchShipments = useCallback(async () => {
    const requestId = ++shipmentListRequestId.current;
    setIsListLoading(true);
    setListError('');
    try {
      const data = await listAdminShipments({
        page,
        limit,
        search,
        shipping_status: selectedStatus || undefined,
        type: selectedType || undefined,
        courier_name: appliedAdvancedFilters.courier_name.trim() || undefined,
        order_id: appliedAdvancedFilters.order_id.trim() || undefined,
        return_request_id: appliedAdvancedFilters.return_request_id.trim() || undefined,
        created_from: appliedAdvancedFilters.created_from || undefined,
        created_to: appliedAdvancedFilters.created_to || undefined,
        sort,
      });
      if (requestId !== shipmentListRequestId.current) return;
      setShipments(data.shipments);
      setTotalItems(data.total_count ?? 0);
      setTotalPages(Math.max(data.total_pages ?? 1, 1));
    } catch (error) {
      if (requestId !== shipmentListRequestId.current) return;
      console.error('Failed to load shipments:', error);
      setShipments([]);
      setTotalItems(0);
      setTotalPages(1);
      setListError(error?.message || 'Unable to load shipments.');
      showToast(error?.message || 'Unable to load shipments.', 'error');
    } finally {
      if (requestId === shipmentListRequestId.current) {
        setIsListLoading(false);
      }
    }
  }, [
    page,
    limit,
    search,
    selectedStatus,
    selectedType,
    appliedAdvancedFilters,
    sort,
    showToast,
  ]);

  useEffect(() => {
    void fetchShipments();
  }, [fetchShipments]);

  useEffect(() => {
    const timeoutId = window.setTimeout(() => {
      const nextSearch = searchInput.trim();
      if (nextSearch !== search) {
        setPage(1);
        setSearch(nextSearch);
      }
    }, 300);
    return () => window.clearTimeout(timeoutId);
  }, [search, searchInput]);

  const lookupShipment = async (type, value) => {
    const trimmedValue = value.trim();
    if (type === 'order' && !ORDER_ID_PATTERN.test(trimmedValue)) {
      showToast('Enter a valid order UUID.', 'error');
      return;
    }
    const shipmentId = Number(trimmedValue);
    if (type === 'shipment' && (!/^\d+$/.test(trimmedValue) || !Number.isSafeInteger(shipmentId) || shipmentId < 1)) {
      showToast('Enter a valid shipment ID.', 'error');
      return;
    }

    setLookupType(type);
    setLookupValue(trimmedValue);
    setIsLoading(true);
    setHasSearched(true);
    setNotFoundLookup(null);
    setShipment(null);
    setHistory([]);
    setIsDetailsOpen(false);
    try {
      const shipmentData = type === 'order'
        ? await getShipmentByOrderId(trimmedValue)
        : await getShipmentById(trimmedValue);
      if (!shipmentData) {
        throw new Error('The shipment lookup returned no shipment data.');
      }

      setShipment(shipmentData);
      setIsDetailsOpen(true);
      try {
        const historyResponse = await getShipmentTrackingHistory(shipmentData.id);
        setHistory(getHistoryItems(historyResponse, shipmentData.history));
      } catch (error) {
        console.error('Failed to load shipment tracking history:', error);
        setHistory(Array.isArray(shipmentData.history) ? shipmentData.history : []);
        showToast('Shipment loaded, but its tracking history could not be refreshed.', 'error');
      }
    } catch (error) {
      if (error?.status === 404) {
        setNotFoundLookup({ type, value: trimmedValue });
      } else {
        console.error('Failed to look up shipment:', error);
        showToast(error?.message || 'Failed to look up shipment.', 'error');
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleLookup = (event) => {
    event.preventDefault();
    void lookupShipment(lookupType, lookupValue);
  };

  const openCreateModal = (orderId = '') => {
    setCreateOrderId(orderId);
    setCourierName('');
    setTrackingNumber('');
    setIsCreateOpen(true);
  };

  const handleCreateShipment = async (event) => {
    event.preventDefault();
    const orderId = createOrderId.trim();
    if (!ORDER_ID_PATTERN.test(orderId)) {
      showToast('Enter a valid order UUID.', 'error');
      return;
    }

    setIsCreating(true);
    try {
      await createShipment({
        order_id: orderId,
        courier_name: courierName.trim() || undefined,
        tracking_number: trackingNumber.trim() || undefined,
      });
      showToast('Shipment created successfully.', 'success');
      setIsCreateOpen(false);
      setLookupType('order');
      setLookupValue(orderId);
      await fetchShipments();
      await lookupShipment('order', orderId);
    } catch (error) {
      console.error('Failed to create shipment:', error);
      showToast(error?.message || 'Failed to create shipment.', 'error');
    } finally {
      setIsCreating(false);
    }
  };

  const openEditTracking = () => {
    setEditCourierName(shipment?.courier_name || '');
    setEditTrackingNumber(shipment?.tracking_number || '');
    setIsDetailsOpen(false);
    setIsEditOpen(true);
  };

  const closeEditTracking = () => {
    setIsEditOpen(false);
    setIsDetailsOpen(Boolean(shipment));
  };

  const handleUpdateTracking = async (event) => {
    event.preventDefault();
    if (!shipment?.id) return;

    setIsSavingTracking(true);
    try {
      await updateShipment(shipment.id, {
        courier_name: editCourierName.trim(),
        tracking_number: editTrackingNumber.trim(),
      });
      showToast('Shipment tracking details updated.', 'success');
      setIsEditOpen(false);
      await fetchShipments();
      await lookupShipment('shipment', String(shipment.id));
    } catch (error) {
      console.error('Failed to update shipment tracking details:', error);
      showToast(error?.message || 'Failed to update tracking details.', 'error');
    } finally {
      setIsSavingTracking(false);
    }
  };

  const openStatusModal = () => {
    const transitions = getShipmentStatusTransitions(shipment?.shipping_status);
    if (!transitions.length) return;
    setTargetStatus(transitions[0]);
    setStatusLocation('');
    setStatusDescription('');
    setIsDetailsOpen(false);
    setIsStatusOpen(true);
  };

  const closeStatusModal = () => {
    setIsStatusOpen(false);
    setIsDetailsOpen(Boolean(shipment));
  };

  const handleUpdateStatus = async (event) => {
    event.preventDefault();
    if (!shipment?.id || !targetStatus) return;
    if (!getShipmentStatusTransitions(shipment.shipping_status).includes(targetStatus)) {
      showToast('That shipment status transition is not allowed.', 'error');
      return;
    }

    setIsUpdatingStatus(true);
    try {
      await updateShippingStatus(shipment.id, {
        status: targetStatus,
        description: statusDescription.trim() || undefined,
        location: statusLocation.trim() || undefined,
      });
      showToast('Shipment status updated.', 'success');
      setIsStatusOpen(false);
      await fetchShipments();
      await lookupShipment('shipment', String(shipment.id));
    } catch (error) {
      console.error('Failed to update shipment status:', error);
      showToast(error?.message || 'Failed to update shipment status.', 'error');
    } finally {
      setIsUpdatingStatus(false);
    }
  };

  const availableTransitions = getShipmentStatusTransitions(shipment?.shipping_status);
  const isCurrentLookupNotFound = notFoundLookup?.type === lookupType
    && notFoundLookup?.value === lookupValue.trim();
  const canCreateForOrder = isCurrentLookupNotFound
    && lookupType === 'order'
    && ORDER_ID_PATTERN.test(lookupValue.trim());

  const handleSearchSubmit = (event) => {
    event.preventDefault();
    setPage(1);
    setSearch(searchInput.trim());
  };

  const handleAdvancedFilterSubmit = (event) => {
    event.preventDefault();
    const { created_from: createdFrom, created_to: createdTo } = advancedFilters;
    if (createdFrom && createdTo && createdFrom > createdTo) {
      showToast('The start date must be on or before the end date.', 'error');
      return;
    }
    const returnRequestId = advancedFilters.return_request_id.trim();
    const parsedReturnRequestId = Number(returnRequestId);
    if (returnRequestId && (
      !/^\d+$/.test(returnRequestId)
      || !Number.isSafeInteger(parsedReturnRequestId)
      || parsedReturnRequestId < 1
      || parsedReturnRequestId > 4294967295
    )) {
      showToast('Return request ID must be a positive number.', 'error');
      return;
    }
    setPage(1);
    setAppliedAdvancedFilters({ ...advancedFilters });
  };

  const handleClearFilters = () => {
    setSearch('');
    setSearchInput('');
    setSelectedStatus('');
    setSelectedType('');
    setSort('newest');
    setAdvancedFilters({
      courier_name: '',
      order_id: '',
      return_request_id: '',
      created_from: '',
      created_to: '',
    });
    setAppliedAdvancedFilters({
      courier_name: '',
      order_id: '',
      return_request_id: '',
      created_from: '',
      created_to: '',
    });
    setLimit(10);
    setPage(1);
  };

  const hasFilters = Boolean(
    search || selectedStatus || selectedType || sort !== 'newest'
      || appliedAdvancedFilters.courier_name
      || appliedAdvancedFilters.order_id
      || appliedAdvancedFilters.return_request_id
      || appliedAdvancedFilters.created_from
      || appliedAdvancedFilters.created_to
  );

  const columns = [
    {
      key: 'id',
      header: 'Shipment',
      render: (id, row) => (
        <div className="admin-shipment-table-primary">
          <strong>#{id}</strong>
          <span>{formatDate(row.created_at)}</span>
        </div>
      ),
    },
    {
      key: 'order_number',
      header: 'Order',
      render: (orderNumber, row) => (
        <Link to={`/admin/orders/${row.order_id}`} className="admin-shipment-order-link">
          {typeof orderNumber === 'string' && orderNumber
            ? orderNumber
            : `Order ${row.order_id?.slice(0, 8) || '—'}`}
        </Link>
      ),
    },
    {
      key: 'type',
      header: 'Type',
      render: (type) => (
        <span className="admin-shipment-type">
          {SHIPMENT_TYPE_LABELS[type] || type || SHIPMENT_TYPE_LABELS.delivery}
        </span>
      ),
    },
    {
      key: 'tracking_number',
      header: 'Courier / Tracking',
      render: (trackingNumber, row) => (
        <div className="admin-shipment-table-primary">
          <strong>{row.courier_name || 'Courier not assigned'}</strong>
          <span>{typeof trackingNumber === 'string' && trackingNumber
            ? trackingNumber
            : 'Tracking not assigned'}</span>
        </div>
      ),
    },
    {
      key: 'shipping_status',
      header: 'Status',
      align: 'center',
      render: (status) => (
        <Badge variant="info">
          {SHIPMENT_STATUS_LABELS[status] || status || 'Unknown'}
        </Badge>
      ),
    },
    {
      key: 'actions',
      header: 'Actions',
      align: 'right',
      render: (_, row) => (
        <Button
          variant="outline"
          size="sm"
          leftIcon={<Eye size={14} />}
          onClick={() => void lookupShipment('shipment', String(row.id))}
        >
          Manage
        </Button>
      ),
    },
  ];

  return (
    <main className="admin-shipments-page">
      <header className="admin-shipments-header">
        <div>
          <span className="admin-shipments-eyebrow">Fulfillment</span>
          <h1>Shipment Management</h1>
          <p>Browse shipments, update delivery details, and review tracking history.</p>
        </div>
        <div className="admin-shipments-header-actions">
          <Button
            type="button"
            variant="outline"
            onClick={() => void fetchShipments()}
            disabled={isListLoading}
            leftIcon={<RefreshCw size={15} className={isListLoading ? 'spin-icon' : ''} />}
          >
            Refresh
          </Button>
          <Button
            type="button"
            variant="primary"
            onClick={() => openCreateModal(lookupType === 'order' ? lookupValue : '')}
            leftIcon={<Package size={16} />}
          >
            Create Shipment
          </Button>
        </div>
      </header>

      <section className="admin-shipment-list-filters" aria-label="Shipment filters">
        <form className="admin-shipment-list-filter-form" onSubmit={handleSearchSubmit}>
          <label className="admin-shipment-list-search">
            <span>Search shipments</span>
            <div className="admin-shipment-search-input">
              <Search size={16} aria-hidden="true" />
              <input
                value={searchInput}
                onChange={(event) => setSearchInput(event.target.value)}
                placeholder="Shipment ID, order number, courier, or tracking number"
              />
            </div>
          </label>
          <label>
            <span>Status</span>
            <select
              value={selectedStatus}
              onChange={(event) => {
                setSelectedStatus(event.target.value);
                setPage(1);
              }}
            >
              <option value="">All statuses</option>
              {Object.entries(SHIPMENT_STATUS_LABELS).map(([value, label]) => (
                <option key={value} value={value}>{label}</option>
              ))}
            </select>
          </label>
          <label>
            <span>Type</span>
            <select
              value={selectedType}
              onChange={(event) => {
                setSelectedType(event.target.value);
                setPage(1);
              }}
            >
              <option value="">All types</option>
              <option value="delivery">Delivery</option>
              <option value="return">Return</option>
              <option value="return_to_customer">Return to customer</option>
            </select>
          </label>
          <Button type="submit" variant="primary" leftIcon={<Search size={15} />}>
            Search
          </Button>
          {hasFilters && (
            <Button type="button" variant="ghost" onClick={handleClearFilters}>
              Clear filters
            </Button>
          )}
        </form>

        <details className="admin-shipment-advanced-filters">
          <summary><Filter size={15} /> Advanced filters and sorting</summary>
          <form className="admin-shipment-advanced-form" onSubmit={handleAdvancedFilterSubmit}>
            <label>
              <span>Courier</span>
              <input
                value={advancedFilters.courier_name}
                onChange={(event) => setAdvancedFilters((current) => ({ ...current, courier_name: event.target.value }))}
                placeholder="Courier name"
              />
            </label>
            <label>
              <span>Order UUID</span>
              <input
                value={advancedFilters.order_id}
                onChange={(event) => setAdvancedFilters((current) => ({ ...current, order_id: event.target.value }))}
                placeholder="Filter by order UUID"
              />
            </label>
            <label>
              <span>Return request ID</span>
              <input
                inputMode="numeric"
                value={advancedFilters.return_request_id}
                onChange={(event) => setAdvancedFilters((current) => ({ ...current, return_request_id: event.target.value }))}
                placeholder="Optional"
              />
            </label>
            <label>
              <span>Created from</span>
              <input
                type="date"
                value={advancedFilters.created_from}
                onChange={(event) => setAdvancedFilters((current) => ({ ...current, created_from: event.target.value }))}
              />
            </label>
            <label>
              <span>Created to</span>
              <input
                type="date"
                value={advancedFilters.created_to}
                onChange={(event) => setAdvancedFilters((current) => ({ ...current, created_to: event.target.value }))}
              />
            </label>
            <label>
              <span>Sort by</span>
              <select value={sort} onChange={(event) => { setSort(event.target.value); setPage(1); }}>
                <option value="newest">Newest first</option>
                <option value="oldest">Oldest first</option>
                <option value="updated_desc">Recently updated</option>
                <option value="updated_asc">Least recently updated</option>
                <option value="status_asc">Status (A-Z)</option>
                <option value="status_desc">Status (Z-A)</option>
              </select>
            </label>
            <label>
              <span>Rows per page</span>
              <select value={limit} onChange={(event) => { setLimit(Number(event.target.value)); setPage(1); }}>
                {[10, 25, 50, 100].map((value) => (
                  <option key={value} value={value}>{value}</option>
                ))}
              </select>
            </label>
            <Button type="submit" variant="secondary">
              Apply advanced filters
            </Button>
          </form>
        </details>
      </section>

      {listError && <p className="admin-shipment-list-error" role="alert">{listError}</p>}

      <DataTable
        columns={columns}
        data={shipments}
        keyField="id"
        isLoading={isListLoading}
        emptyMessage={listError || 'No shipments match the selected filters.'}
        emptyAction={listError ? null : (
          <Button type="button" variant="outline" onClick={() => openCreateModal()}>
            Create Shipment
          </Button>
        )}
        className="admin-shipment-data-table"
        pagination={{
          page,
          totalPages,
          totalItems,
          limit,
          onPageChange: setPage,
        }}
      />

      <details className="admin-shipment-manual-lookup">
        <summary>Open a shipment directly by its ID or order UUID</summary>
        <section className="admin-shipment-lookup-card" aria-labelledby="shipment-lookup-title">
          <div className="admin-shipment-section-heading">
            <div className="admin-shipment-icon"><Search size={18} /></div>
            <div>
              <h2 id="shipment-lookup-title">Direct lookup</h2>
              <p>Use this when you already have the shipment ID or order UUID.</p>
            </div>
          </div>
          <form className="admin-shipment-lookup-form" onSubmit={handleLookup}>
            <label>
              <span>Search by</span>
              <select value={lookupType} onChange={(event) => setLookupType(event.target.value)}>
                <option value="order">Order UUID</option>
                <option value="shipment">Shipment ID</option>
              </select>
            </label>
            <label className="admin-shipment-lookup-value">
              <span>{lookupType === 'order' ? 'Order UUID' : 'Shipment ID'}</span>
              <input
                value={lookupValue}
                onChange={(event) => setLookupValue(event.target.value)}
                placeholder={lookupType === 'order' ? 'e.g. 123e4567-e89b-12d3-a456-426614174000' : 'e.g. 42'}
                autoComplete="off"
              />
            </label>
            <Button type="submit" variant="primary" disabled={isLoading} leftIcon={<Search size={15} />}>
              {isLoading ? 'Searching...' : 'Search'}
            </Button>
          </form>
        </section>
      </details>

      {isLoading && <p className="admin-shipment-detail-loading" role="status">Loading shipment details...</p>}

      <Modal
        isOpen={isDetailsOpen && Boolean(shipment)}
        onClose={() => setIsDetailsOpen(false)}
        title={`Shipment #${shipment?.id ?? ''}`}
        size="lg"
      >
        {shipment && (
          <section className="admin-shipment-result-card" aria-labelledby="shipment-result-title">
            <div className="admin-shipment-result-header">
              <div className="admin-shipment-result-title">
                <div className="admin-shipment-icon"><Truck size={19} /></div>
                <div>
                  <span className="admin-shipment-meta">Shipment #{shipment.id}</span>
                  <h2 id="shipment-result-title">
                    {SHIPMENT_TYPE_LABELS[shipment.type] || shipment.type || SHIPMENT_TYPE_LABELS.delivery} shipment
                  </h2>
                </div>
              </div>
              <Badge variant="info">
                {SHIPMENT_STATUS_LABELS[shipment.shipping_status] || shipment.shipping_status || 'Unknown'}
              </Badge>
            </div>

            <div className="admin-shipment-details-grid">
              <div>
                <span className="admin-shipment-detail-label">Order</span>
                {shipment.order_id ? (
                  <Link to={`/admin/orders/${shipment.order_id}`} className="admin-shipment-order-link">
                    View order
                  </Link>
                ) : <span className="admin-shipment-detail-value">—</span>}
              </div>
              <div>
                <span className="admin-shipment-detail-label">Courier</span>
                <span className="admin-shipment-detail-value">{shipment.courier_name || 'Not assigned'}</span>
              </div>
              <div>
                <span className="admin-shipment-detail-label">Tracking number</span>
                <span className="admin-shipment-detail-value">{shipment.tracking_number || 'Not assigned'}</span>
              </div>
              <div>
                <span className="admin-shipment-detail-label">Shipped</span>
                <span className="admin-shipment-detail-value">{shipment.shipped_at ? formatDate(shipment.shipped_at) : '—'}</span>
              </div>
              <div>
                <span className="admin-shipment-detail-label">Delivered</span>
                <span className="admin-shipment-detail-value">{shipment.delivered_at ? formatDate(shipment.delivered_at) : '—'}</span>
              </div>
            </div>

            <div className="admin-shipment-actions">
              <Button type="button" variant="outline" onClick={openEditTracking}>
                Update Tracking Details
              </Button>
              {availableTransitions.length > 0 && (
                <Button type="button" variant="primary" onClick={openStatusModal}>
                  Update Status
                </Button>
              )}
            </div>

            <div className="admin-shipment-history">
              <div className="admin-shipment-history-heading">
                <div>
                  <Clock3 size={16} />
                  <h3>Tracking history</h3>
                </div>
                <span>{history.length} {history.length === 1 ? 'update' : 'updates'}</span>
              </div>
              {history.length ? (
                <ol className="admin-shipment-history-list">
                  {history.map((item, index) => (
                    <li key={item.id || `${item.created_at}-${index}`} className="admin-shipment-history-item">
                      <span className="admin-shipment-history-dot" />
                      <div className="admin-shipment-history-content">
                        <strong>{SHIPMENT_STATUS_LABELS[item.shipping_status] || item.shipping_status}</strong>
                        {item.location && <span>{item.location}</span>}
                        {item.description && <p>{item.description}</p>}
                        <time>{item.created_at ? formatDate(item.created_at) : 'Date unavailable'}</time>
                      </div>
                    </li>
                  ))}
                </ol>
              ) : (
                <p className="admin-shipment-history-empty">No tracking updates have been recorded.</p>
              )}
            </div>
          </section>
        )}
      </Modal>

      {canCreateForOrder && (
        <section className="admin-shipment-not-found" role="status">
          <Package size={24} />
          <div>
            <h2>No shipment found for this order</h2>
            <p>Create a shipment if the order is ready for fulfillment.</p>
          </div>
          <Button type="button" variant="primary" onClick={() => openCreateModal(lookupValue.trim())}>
            Create Shipment
          </Button>
        </section>
      )}

      {hasSearched && isCurrentLookupNotFound && !canCreateForOrder && !isLoading && (
        <section className="admin-shipment-not-found" role="status">
          <Package size={24} />
          <div>
            <h2>No shipment found</h2>
            <p>Check the identifier and try again.</p>
          </div>
        </section>
      )}

      <Modal isOpen={isCreateOpen} onClose={() => setIsCreateOpen(false)} title="Create Shipment">
        <form className="admin-shipment-modal-form" onSubmit={handleCreateShipment}>
          <label>
            <span>Order UUID *</span>
            <input
              value={createOrderId}
              onChange={(event) => setCreateOrderId(event.target.value)}
              placeholder="Paste the order UUID"
              required
            />
          </label>
          <label>
            <span>Courier name</span>
            <input value={courierName} onChange={(event) => setCourierName(event.target.value)} />
          </label>
          <label>
            <span>Tracking number</span>
            <input value={trackingNumber} onChange={(event) => setTrackingNumber(event.target.value)} />
          </label>
          <div className="admin-shipment-modal-actions">
            <Button type="button" variant="outline" onClick={() => setIsCreateOpen(false)}>Cancel</Button>
            <Button type="submit" variant="primary" disabled={isCreating}>
              {isCreating ? 'Creating...' : 'Create Shipment'}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal isOpen={isEditOpen} onClose={closeEditTracking} title="Update Tracking Details">
        <form className="admin-shipment-modal-form" onSubmit={handleUpdateTracking}>
          <label>
            <span>Courier name</span>
            <input value={editCourierName} onChange={(event) => setEditCourierName(event.target.value)} />
          </label>
          <label>
            <span>Tracking number</span>
            <input value={editTrackingNumber} onChange={(event) => setEditTrackingNumber(event.target.value)} />
          </label>
          <div className="admin-shipment-modal-actions">
            <Button type="button" variant="outline" onClick={closeEditTracking}>Cancel</Button>
            <Button type="submit" variant="primary" disabled={isSavingTracking}>
              {isSavingTracking ? 'Saving...' : 'Save Details'}
            </Button>
          </div>
        </form>
      </Modal>

      <Modal isOpen={isStatusOpen} onClose={closeStatusModal} title="Update Shipment Status">
        <form className="admin-shipment-modal-form" onSubmit={handleUpdateStatus}>
          <label>
            <span>New status *</span>
            <select value={targetStatus} onChange={(event) => setTargetStatus(event.target.value)} required>
              {availableTransitions.map((status) => (
                <option key={status} value={status}>{SHIPMENT_STATUS_LABELS[status] || status}</option>
              ))}
            </select>
          </label>
          <label>
            <span>Location</span>
            <input value={statusLocation} onChange={(event) => setStatusLocation(event.target.value)} />
          </label>
          <label>
            <span>Description</span>
            <textarea
              value={statusDescription}
              onChange={(event) => setStatusDescription(event.target.value)}
              rows={3}
            />
          </label>
          <div className="admin-shipment-modal-actions">
            <Button type="button" variant="outline" onClick={closeStatusModal}>Cancel</Button>
            <Button type="submit" variant="primary" disabled={isUpdatingStatus || !targetStatus}>
              {isUpdatingStatus ? 'Updating...' : 'Update Status'}
            </Button>
          </div>
        </form>
      </Modal>
    </main>
  );
}
