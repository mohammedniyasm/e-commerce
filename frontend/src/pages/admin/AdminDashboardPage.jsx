import { useEffect, useMemo, useState } from 'react';
import { Link } from 'react-router-dom';
import {
  AlertTriangle,
  ArrowUpRight,
  BarChart3,
  ChevronRight,
  CreditCard,
  IndianRupee,
  Package,
  RotateCcw,
  RotateCw,
  ShoppingBag,
  TicketPercent,
  TrendingUp,
  Truck,
  Users,
} from 'lucide-react';
import {
  getAdminRecentOrders,
  getDashboardCustomerStats,
  getDashboardOrderStats,
  getDashboardOverview,
  getDashboardSales,
  getDashboardTopProducts,
} from '../../api/adminService';
import { Badge } from '../../components/common/Badge';
import { formatCurrency, formatDate, formatOrderStatus } from '../../utils/formatters';
import './AdminDashboardPage.css';

export function fillMissingSalesBuckets(salesData, filter) {
  if (!Array.isArray(salesData) || salesData.length === 0) return [];
  if (!filter?.start_date || !filter?.end_date || !['day', 'month'].includes(filter.group_by)) {
    return salesData;
  }

  const bucketKey = (date) => {
    const dateText = date instanceof Date ? date.toISOString().slice(0, 10) : String(date || '').slice(0, 10);
    if (!/^\d{4}-\d{2}-\d{2}$/.test(dateText)) return null;
    return filter.group_by === 'month' ? dateText.slice(0, 7) : dateText;
  };
  const salesByBucket = new Map();
  for (const row of salesData) {
    const key = bucketKey(row.date);
    if (key) salesByBucket.set(key, row);
  }

  const cursor = new Date(`${filter.start_date}T00:00:00.000Z`);
  const lastDate = new Date(`${filter.end_date}T00:00:00.000Z`);
  if (Number.isNaN(cursor.getTime()) || Number.isNaN(lastDate.getTime()) || cursor > lastDate) {
    return salesData;
  }
  if (filter.group_by === 'month') {
    cursor.setUTCDate(1);
    lastDate.setUTCDate(1);
  }

  const buckets = [];
  while (cursor <= lastDate) {
    const dateText = cursor.toISOString().slice(0, 10);
    const key = filter.group_by === 'month' ? dateText.slice(0, 7) : dateText;
    const row = salesByBucket.get(key);
    buckets.push({
      date: row?.date ?? `${filter.group_by === 'month' ? `${key}-01` : key}T00:00:00.000Z`,
      total_revenue: Number(row?.total_revenue) || 0,
      total_orders: Number(row?.total_orders) || 0,
      hasActivity: Boolean(row),
    });

    if (filter.group_by === 'month') {
      cursor.setUTCMonth(cursor.getUTCMonth() + 1);
    } else {
      cursor.setUTCDate(cursor.getUTCDate() + 1);
    }
  }

  return buckets;
}

export function AdminDashboardPage() {
  const [overview, setOverview] = useState(null);
  const [salesData, setSalesData] = useState([]);
  const [orderStats, setOrderStats] = useState(null);
  const [customerStats, setCustomerStats] = useState(null);
  const [topProducts, setTopProducts] = useState([]);
  const [recentOrders, setRecentOrders] = useState([]);

  const [timeRange, setTimeRange] = useState('7d'); // '7d' | '30d' | 'year'
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [error, setError] = useState(null);
  const [hoveredPoint, setHoveredPoint] = useState(null);

  // Calculate start/end date for sales query
  const salesFilter = useMemo(() => {
    const end = new Date();
    const start = new Date();

    if (timeRange === '7d') {
      start.setDate(end.getDate() - 7);
      return {
        group_by: 'day',
        start_date: start.toISOString().split('T')[0],
        end_date: end.toISOString().split('T')[0],
      };
    }
    if (timeRange === '30d') {
      start.setDate(end.getDate() - 30);
      return {
        group_by: 'day',
        start_date: start.toISOString().split('T')[0],
        end_date: end.toISOString().split('T')[0],
      };
    }
    // Year
    start.setFullYear(end.getFullYear() - 1);
    return {
      group_by: 'month',
      start_date: start.toISOString().split('T')[0],
      end_date: end.toISOString().split('T')[0],
    };
  }, [timeRange]);

  async function loadDashboardData(showFullLoading = true) {
    if (showFullLoading) setIsLoading(true);
    else setIsRefreshing(true);
    setError(null);

    try {
      const [
        overviewResult,
        salesResult,
        orderStatsResult,
        customerStatsResult,
        topProductsResult,
        recentOrdersResult,
      ] = await Promise.allSettled([
        getDashboardOverview(),
        getDashboardSales(salesFilter),
        getDashboardOrderStats(),
        getDashboardCustomerStats(),
        getDashboardTopProducts(5),
        getAdminRecentOrders({ limit: 5, page: 1 }),
      ]);

      if (overviewResult.status === 'fulfilled' && overviewResult.value) {
        console.log(orderStatsResult)
        setOverview(overviewResult.value);
      }
      if (salesResult.status === 'fulfilled' && Array.isArray(salesResult.value)) {
        setSalesData(salesResult.value);
      }
      if (orderStatsResult.status === 'fulfilled' && orderStatsResult.value) {
        setOrderStats(orderStatsResult.value);
      }
      if (customerStatsResult.status === 'fulfilled' && customerStatsResult.value) {
        setCustomerStats(customerStatsResult.value);
      }
      if (topProductsResult.status === 'fulfilled' && Array.isArray(topProductsResult.value)) {
        setTopProducts(topProductsResult.value);
      }
      if (recentOrdersResult.status === 'fulfilled' && recentOrdersResult.value?.orders) {
        setRecentOrders(recentOrdersResult.value.orders);
      }
    } catch (err) {
      setError(err?.message || 'Failed to load dashboard analytics');
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
    }
  }

  useEffect(() => {
    loadDashboardData(true);
  }, []);

  // When time range changes, update sales query specifically
  useEffect(() => {
    let isCurrent = true;
    getDashboardSales(salesFilter)
      .then((data) => {
        if (isCurrent && Array.isArray(data)) {
          setSalesData(data);
        }
      })
      .catch(() => { });

    return () => {
      isCurrent = false;
    };
  }, [salesFilter]);

  // Aggregate Sales Summary
  const { totalPeriodRevenue, totalPeriodOrders, avgOrderValue } = useMemo(() => {
    if (!salesData.length) {
      return { totalPeriodRevenue: 0, totalPeriodOrders: 0, avgOrderValue: 0 };
    }
    const rev = salesData.reduce((acc, curr) => acc + (Number(curr.total_revenue) || 0), 0);
    const ord = salesData.reduce((acc, curr) => acc + (Number(curr.total_orders) || 0), 0);
    const aov = ord > 0 ? rev / ord : 0;
    return {
      totalPeriodRevenue: rev,
      totalPeriodOrders: ord,
      avgOrderValue: aov,
    };
  }, [salesData]);

  // SVG Chart Geometry Calculation
  const chartGeometry = useMemo(() => {
    const chartSalesData = fillMissingSalesBuckets(salesData, salesFilter);
    if (chartSalesData.length === 0) return null;

    const width = 800;
    const height = 200;
    const paddingX = 40;
    const paddingY = 25;

    const revenues = chartSalesData.map((d) => Number(d.total_revenue) || 0);
    const maxRevenue = Math.max(...revenues, 100);
    const minRevenue = 0;

    const points = chartSalesData.map((d, index) => {
      const x =
        chartSalesData.length === 1
          ? width / 2
          : paddingX + (index / (chartSalesData.length - 1)) * (width - 2 * paddingX);
      const revenue = Number(d.total_revenue) || 0;
      const y =
        height -
        paddingY -
        ((revenue - minRevenue) / (maxRevenue - minRevenue || 1)) * (height - 2 * paddingY);
      return {
        x,
        y,
        date: d.date,
        revenue,
        orders: d.total_orders || 0,
        hasActivity: d.hasActivity !== false,
      };
    });

    const linePath = points.reduce((acc, p, i) => `${acc} ${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`, '');
    const areaPath = `${linePath} L ${points[points.length - 1].x} ${height - paddingY} L ${points[0].x} ${height - paddingY} Z`;

    return {
      width,
      height,
      points,
      linePath,
      areaPath,
      maxRevenue,
      paddingY,
    };
  }, [salesData, salesFilter]);

  return (
    <div className="admin-dashboard">
      {/* Header */}
      <div className="admin-dashboard-header">
        <div className="admin-header-title-group">
          <h1 className="admin-dashboard-title">Dashboard Overview</h1>
          <p className="admin-dashboard-subtitle">
            Real-time operations summary, sales trends, and order performance.
          </p>
        </div>

        <button
          type="button"
          className="admin-refresh-btn"
          onClick={() => loadDashboardData(false)}
          disabled={isLoading || isRefreshing}
          aria-label="Refresh dashboard metrics"
        >
          <RotateCw size={14} className={isRefreshing ? 'spin-icon' : ''} />
          <span>{isRefreshing ? 'Refreshing...' : 'Refresh Metrics'}</span>
        </button>
      </div>

      {error && (
        <div className="admin-error-card" role="alert">
          <AlertTriangle size={24} color="#dc2626" />
          <h3 className="admin-error-title">Unable to Load Complete Analytics</h3>
          <p className="admin-error-desc">{error}</p>
          <button
            type="button"
            className="admin-refresh-btn"
            onClick={() => loadDashboardData(true)}
          >
            Retry Fetching Data
          </button>
        </div>
      )}

      {/* Operations Quick Navigation Tabs */}
      <nav className="admin-operations-tabs" aria-label="Operations Quick Navigation">
        <Link to="/admin/orders" className="admin-op-tab">
          <ShoppingBag size={15} />
          <span>Orders</span>
          {overview?.pending_orders > 0 && (
            <span className="admin-op-tab-badge">{overview.pending_orders}</span>
          )}
        </Link>
        <Link to="/admin/payments" className="admin-op-tab">
          <CreditCard size={15} />
          <span>Payments</span>
        </Link>
        <Link to="/admin/customers" className="admin-op-tab">
          <Users size={15} />
          <span>Customers</span>
        </Link>
        <Link to="/admin/products" className="admin-op-tab">
          <Package size={15} />
          <span>Products</span>
        </Link>
        <Link to="/admin/returns" className="admin-op-tab">
          <RotateCcw size={15} />
          <span>Returns</span>
        </Link>
        <Link to="/admin/shipments" className="admin-op-tab">
          <Truck size={15} />
          <span>Shipments</span>
        </Link>
        <Link to="/admin/coupons" className="admin-op-tab">
          <TicketPercent size={15} />
          <span>Coupons</span>
        </Link>
        <Link to="/admin/reports" className="admin-op-tab">
          <BarChart3 size={15} />
          <span>Reports</span>
        </Link>
      </nav>

      {/* KPI Cards */}
      <section className="admin-kpi-grid" aria-label="Key Performance Indicators">
        {/* Card 1: Total Orders */}
        <Link to="/admin/orders" className="admin-kpi-card admin-kpi-card--clickable">
          <div className="admin-kpi-header">
            <span className="admin-kpi-label">Total Orders</span>
            <div className="admin-kpi-icon-wrapper orders">
              <ShoppingBag size={20} />
            </div>
          </div>
          <div className="admin-kpi-body">
            <span className="admin-kpi-value">
              {Number(orderStats?.total_orders ?? overview?.total_orders ?? 0).toLocaleString()}
            </span>
            <div className="admin-kpi-detail">
              <span className="admin-kpi-trend positive">
                <TrendingUp size={12} />
                <span>Active Store</span>
              </span>
              <span>{orderStats?.pending_orders ?? overview?.pending_orders ?? 0} pending processing</span>
            </div>
          </div>
        </Link>

        {/* Card 2: Total Revenue */}
        <Link to="/admin/payments" className="admin-kpi-card admin-kpi-card--clickable">
          <div className="admin-kpi-header">
            <span className="admin-kpi-label">Total Revenue</span>
            <div className="admin-kpi-icon-wrapper revenue">
              <IndianRupee size={20} />
            </div>
          </div>
          <div className="admin-kpi-body">
            <span className="admin-kpi-value">
              {formatCurrency(overview?.total_revenue || 0)}
            </span>
            <div className="admin-kpi-detail">
              <span className="admin-kpi-trend positive">
                <ArrowUpRight size={12} />
                <span>Gross</span>
              </span>
              <span>{orderStats?.delivered_orders ?? overview?.delivered_orders ?? 0} orders delivered</span>
            </div>
          </div>
        </Link>

        {/* Card 3: Total Customers */}
        <Link to="/admin/customers" className="admin-kpi-card admin-kpi-card--clickable">
          <div className="admin-kpi-header">
            <span className="admin-kpi-label">Customers</span>
            <div className="admin-kpi-icon-wrapper customers">
              <Users size={20} />
            </div>
          </div>
          <div className="admin-kpi-body">
            <span className="admin-kpi-value">
              {overview?.total_customers != null
                ? Number(overview.total_customers).toLocaleString()
                : '0'}
            </span>
            <div className="admin-kpi-detail">
              <span>{customerStats?.verified_customers || overview?.total_customers || 0} verified accounts</span>
            </div>
          </div>
        </Link>

        {/* Card 4: Total Products */}
        <Link to="/admin/products" className="admin-kpi-card admin-kpi-card--clickable">
          <div className="admin-kpi-header">
            <span className="admin-kpi-label">Catalog Products</span>
            <div className="admin-kpi-icon-wrapper products">
              <Package size={20} />
            </div>
          </div>
          <div className="admin-kpi-body">
            <span className="admin-kpi-value">
              {overview?.total_products != null
                ? Number(overview.total_products).toLocaleString()
                : '0'}
            </span>
            <div className="admin-kpi-detail">
              <span>Published store inventory</span>
            </div>
          </div>
        </Link>
      </section>

      {/* Revenue Analytics Chart Card */}
      <section className="admin-chart-card" aria-label="Revenue Analytics Chart">
        <div className="admin-chart-header">
          <div className="admin-chart-title-group">
            <h2 className="admin-chart-title">Revenue Analytics</h2>
            <p className="admin-chart-subtitle">Gross sales timeline across recorded transactions</p>
          </div>

          <div className="admin-chart-controls" role="group" aria-label="Time period range">
            <button
              type="button"
              className={`admin-chart-btn ${timeRange === '7d' ? 'active' : ''}`}
              onClick={() => setTimeRange('7d')}
            >
              7 Days
            </button>
            <button
              type="button"
              className={`admin-chart-btn ${timeRange === '30d' ? 'active' : ''}`}
              onClick={() => setTimeRange('30d')}
            >
              30 Days
            </button>
            <button
              type="button"
              className={`admin-chart-btn ${timeRange === 'year' ? 'active' : ''}`}
              onClick={() => setTimeRange('year')}
            >
              1 Year
            </button>
          </div>
        </div>

        <div className="admin-chart-metrics">
          <div className="admin-metric-item">
            <span className="admin-metric-label">Period Revenue</span>
            <span className="admin-metric-value">{formatCurrency(totalPeriodRevenue)}</span>
          </div>
          <div className="admin-metric-item">
            <span className="admin-metric-label">Period Orders</span>
            <span className="admin-metric-value">{totalPeriodOrders.toLocaleString()}</span>
          </div>
          <div className="admin-metric-item">
            <span className="admin-metric-label">Avg. Order Value</span>
            <span className="admin-metric-value">{formatCurrency(avgOrderValue)}</span>
          </div>
        </div>

        {/* Chart Visualization */}
        <div className="admin-chart-container">
          {chartGeometry && chartGeometry.points.length > 0 ? (
            <>
              {hoveredPoint && (
                <div className="chart-tooltip" role="status">
                  <strong>{formatDate(hoveredPoint.date)}</strong>
                  <span>Revenue: {formatCurrency(hoveredPoint.revenue)}</span>
                  <span>Orders: {hoveredPoint.orders}</span>
                </div>
              )}

              <svg
                viewBox={`0 0 ${chartGeometry.width} ${chartGeometry.height}`}
                className="admin-svg-chart"
                preserveAspectRatio="none"
              >
                <defs>
                  <linearGradient id="revenueGradient" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="0%" stopColor="#0f172a" stopOpacity="0.25" />
                    <stop offset="100%" stopColor="#0f172a" stopOpacity="0.0" />
                  </linearGradient>
                </defs>

                {/* Grid lines */}
                <line
                  x1="40"
                  y1={chartGeometry.paddingY}
                  x2={chartGeometry.width - 40}
                  y2={chartGeometry.paddingY}
                  className="chart-grid-line"
                />
                <line
                  x1="40"
                  y1={chartGeometry.height / 2}
                  x2={chartGeometry.width - 40}
                  y2={chartGeometry.height / 2}
                  className="chart-grid-line"
                />
                <line
                  x1="40"
                  y1={chartGeometry.height - chartGeometry.paddingY}
                  x2={chartGeometry.width - 40}
                  y2={chartGeometry.height - chartGeometry.paddingY}
                  className="chart-grid-line"
                />

                {/* Filled area */}
                <path d={chartGeometry.areaPath} className="chart-area-fill" />

                {/* Line */}
                <path d={chartGeometry.linePath} className="chart-line" />

                {/* Data points */}
                {chartGeometry.points.filter((pt) => pt.hasActivity).map((pt, i) => (
                  <circle
                    key={i}
                    cx={pt.x}
                    cy={pt.y}
                    r={hoveredPoint?.x === pt.x ? 5 : 3.5}
                    className={`chart-dot ${hoveredPoint?.x === pt.x ? 'active' : ''}`}
                    onMouseEnter={() => setHoveredPoint(pt)}
                    onMouseLeave={() => setHoveredPoint(null)}
                    aria-label={`${formatDate(pt.date)}: ${formatCurrency(pt.revenue)}`}
                  />
                ))}
              </svg>
            </>
          ) : (
            <div className="admin-empty-state">
              No sales activity recorded for this selected time window.
            </div>
          )}
        </div>
      </section>

      {/* Two Column Section: Recent Orders & Top Selling Products */}
      <div className="admin-dashboard-split">
        {/* Left Column: Recent Orders Table */}
        <section className="admin-card" aria-label="Recent Orders">
          <div className="admin-card-header">
            <h2 className="admin-card-title">Recent Orders</h2>
            <Link to="/admin/orders" className="admin-card-action">
              <span>View All</span>
              <ChevronRight size={14} />
            </Link>
          </div>

          <div className="admin-table-wrapper">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Order #</th>
                  <th>Customer</th>
                  <th>Date</th>
                  <th>Status</th>
                  <th style={{ textAlign: 'right' }}>Amount</th>
                </tr>
              </thead>
              <tbody>
                {recentOrders && recentOrders.length > 0 ? (
                  recentOrders.map((order) => {
                    const statusMeta = formatOrderStatus(order.order_status || order.status);
                    const orderNumber = order.order_number || String(order.id).padStart(5, '0');
                    const customerName =
                      order.shipping_address?.full_name ||
                      order.user?.email ||
                      (order.user_id ? `User #${order.user_id}` : 'Customer');
                    const initials = customerName.slice(0, 2).toUpperCase();

                    return (
                      <tr key={order.id}>
                        <td data-label="Order #">
                          <span className="admin-order-id" title={`#${orderNumber}`}>
                            #{orderNumber}
                          </span>
                        </td>
                        <td data-label="Customer">
                          <div className="admin-customer-cell">
                            <div className="admin-customer-avatar">{initials}</div>
                            <span className="admin-customer-name">{customerName}</span>
                          </div>
                        </td>
                        <td data-label="Date">{formatDate(order.created_at)}</td>
                        <td data-label="Status">
                          <Badge variant={statusMeta.variant} size="sm">
                            {statusMeta.label}
                          </Badge>
                        </td>
                        <td data-label="Amount" className="admin-amount-cell">
                          {formatCurrency(order.total_amount || order.grand_total || 0)}
                        </td>
                      </tr>
                    );
                  })
                ) : (
                  <tr>
                    <td colSpan={5} className="admin-empty-state">
                      No recent customer orders have been placed yet.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </section>

        {/* Right Column: Top Products */}
        <section className="admin-card" aria-label="Top Selling Products">
          <div className="admin-card-header">
            <h2 className="admin-card-title">Top Selling Products</h2>
            <Link to="/admin/products" className="admin-card-action">
              <span>View Catalog</span>
              <ChevronRight size={14} />
            </Link>
          </div>

          <div className="admin-products-list">
            {topProducts && topProducts.length > 0 ? (
              topProducts.map((prod, idx) => (
                <div key={prod.product_id || idx} className="admin-product-row">
                  <div className="admin-product-info-group">
                    <span className={`admin-product-rank rank-${idx + 1}`}>#{idx + 1}</span>
                    <div className="admin-product-details">
                      <span className="admin-product-name" title={prod.product_name}>
                        {prod.product_name || `Product #${prod.product_id}`}
                      </span>
                      <span className="admin-product-sales-sub">
                        {prod.units_sold || 0} units sold ({prod.total_orders || 0} orders)
                      </span>
                    </div>
                  </div>
                  <span className="admin-product-revenue">
                    {formatCurrency(prod.revenue || 0)}
                  </span>
                </div>
              ))
            ) : (
              <div className="admin-empty-state">
                No product performance metrics recorded yet.
              </div>
            )}
          </div>
        </section>
      </div>

      {/* Secondary Row: Order Distribution & Customer Health */}
      <div className="admin-breakdown-grid">
        {/* Order Status Distribution */}
        <section className="admin-card" aria-label="Order Status Distribution">
          <div className="admin-card-header">
            <h2 className="admin-card-title">Order Status Distribution</h2>
          </div>
          <div className="admin-status-bars">
            {orderStats && orderStats.total_orders > 0 ? (
              [
                { label: 'Delivered', count: orderStats.delivered_orders, class: 'delivered' },
                { label: 'Confirmed', count: orderStats.confirmed_orders, class: 'confirmed' },
                { label: 'Processing', count: orderStats.processing_orders, class: 'processing' },
                { label: 'Shipped', count: orderStats.shipped_orders, class: 'shipped' },
                { label: 'Pending', count: orderStats.pending_orders, class: 'pending' },
                { label: 'Cancelled', count: orderStats.cancelled_orders, class: 'cancelled' },
              ].map((item) => {
                const percentage =
                  orderStats.total_orders > 0
                    ? Math.round(((item.count || 0) / orderStats.total_orders) * 100)
                    : 0;
                return (
                  <div key={item.label} className="admin-status-item">
                    <div className="admin-status-meta">
                      <span className="admin-status-name">{item.label}</span>
                      <span className="admin-status-count">
                        {item.count || 0} ({percentage}%)
                      </span>
                    </div>
                    <div className="admin-progress-track">
                      <div
                        className={`admin-progress-fill ${item.class}`}
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="admin-empty-state">
                No order statistics available to display.
              </div>
            )}
          </div>
        </section>

        {/* Customer Account Health */}
        <section className="admin-card" aria-label="Customer Account Health">
          <div className="admin-card-header">
            <h2 className="admin-card-title">Customer Account Health</h2>
          </div>
          <div className="admin-status-bars">
            {customerStats && customerStats.total_customers > 0 ? (
              [
                {
                  label: 'Verified Accounts',
                  count: customerStats.verified_customers,
                  class: 'delivered',
                },
                {
                  label: 'Unverified Accounts',
                  count: customerStats.unverified_customers,
                  class: 'pending',
                },
                {
                  label: 'Blocked / Restricted',
                  count: customerStats.blocked_customers,
                  class: 'cancelled',
                },
              ].map((item) => {
                const percentage =
                  customerStats.total_customers > 0
                    ? Math.round(((item.count || 0) / customerStats.total_customers) * 100)
                    : 0;
                return (
                  <div key={item.label} className="admin-status-item">
                    <div className="admin-status-meta">
                      <span className="admin-status-name">{item.label}</span>
                      <span className="admin-status-count">
                        {item.count || 0} ({percentage}%)
                      </span>
                    </div>
                    <div className="admin-progress-track">
                      <div
                        className={`admin-progress-fill ${item.class}`}
                        style={{ width: `${percentage}%` }}
                      />
                    </div>
                  </div>
                );
              })
            ) : (
              <div className="admin-empty-state">
                No customer account telemetry recorded.
              </div>
            )}
          </div>
        </section>
      </div>
    </div>
  );
}
