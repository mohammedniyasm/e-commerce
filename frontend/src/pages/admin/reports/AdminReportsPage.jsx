import { useEffect, useState, useCallback } from 'react';
import {
  BarChart3,
  TrendingUp,
  Download,
  Calendar,
  FileSpreadsheet,
  FileText,
  RefreshCw,
  ShoppingBag,
  DollarSign,
  Package,
} from 'lucide-react';
import {
  getSalesReport,
  getOrdersReport,
  getProductsReport,
  downloadReportFile,
} from '../../../api/adminOperationsService';
import { Button } from '../../../components/common/Button';
import { DataTable } from '../../../components/common/DataTable';
import { useToast } from '../../../context/ToastContext';
import { formatCurrency, formatDate } from '../../../utils/formatters';
import './AdminReportsPage.css';

export function AdminReportsPage() {
  const { showToast } = useToast();

  const [activeReport, setActiveReport] = useState('sales'); // 'sales' | 'orders' | 'products'
  const [groupBy, setGroupBy] = useState('day'); // 'day' | 'week' | 'month'
  const [startDate, setStartDate] = useState('');
  const [endDate, setEndDate] = useState('');

  const [reportData, setReportData] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isExporting, setIsExporting] = useState(false);

  // Initialize with last 30 days
  useEffect(() => {
    const end = new Date();
    const start = new Date();
    start.setDate(end.getDate() - 30);

    setEndDate(end.toISOString().split('T')[0]);
    setStartDate(start.toISOString().split('T')[0]);
  }, []);

  const fetchReport = useCallback(async () => {
    if (!startDate || !endDate) return;
    setIsLoading(true);
    try {
      const params = {
        group_by: groupBy,
        start_date: startDate,
        end_date: endDate,
      };

      let data = null;
      if (activeReport === 'sales') {
        data = await getSalesReport(params);
      } else if (activeReport === 'orders') {
        data = await getOrdersReport(params);
      } else if (activeReport === 'products') {
        data = await getProductsReport(params);
      }

      setReportData(data);
    } catch (err) {
      console.error('Failed to load report data:', err);
      showToast(err?.response?.data?.message || 'Failed to load report data', 'error');
    } finally {
      setIsLoading(false);
    }
  }, [activeReport, groupBy, startDate, endDate, showToast]);

  useEffect(() => {
    fetchReport();
  }, [fetchReport]);

  const handleExport = async (format) => {
    setIsExporting(true);
    try {
      let endpoint = '';
      let filename = `drape_${activeReport}_report_${startDate}_to_${endDate}.${format === 'pdf' ? 'pdf' : 'xlsx'}`;

      if (activeReport === 'sales') {
        endpoint = `/admin/reports/sales/export/${format}`;
      } else if (activeReport === 'orders') {
        endpoint = `/admin/reports/orders/export/${format}`;
      } else if (activeReport === 'products') {
        endpoint = `/admin/reports/products/export/${format}`;
      }

      const params = {
        group_by: groupBy,
        start_date: startDate,
        end_date: endDate,
      };

      await downloadReportFile(endpoint, params, filename);
      showToast(`Exported ${activeReport} report as ${format.toUpperCase()}`, 'success');
    } catch (err) {
      console.error('Failed to export report:', err);
      showToast(err?.response?.data?.message || 'Failed to download report', 'error');
    } finally {
      setIsExporting(false);
    }
  };

  // Preset Date Ranges
  const applyPreset = (days) => {
    const end = new Date();
    const start = new Date();
    start.setDate(end.getDate() - days);
    setEndDate(end.toISOString().split('T')[0]);
    setStartDate(start.toISOString().split('T')[0]);
  };

  // Data table columns depending on active report
  const getColumns = () => {
    if (activeReport === 'sales') {
      return [
        { key: 'date', header: 'Period / Date' },
        {
          key: 'orders_count',
          header: 'Total Orders',
          align: 'center',
          render: (row) => row.orders_count || row.total_orders || 0,
        },
        {
          key: 'revenue',
          header: 'Revenue',
          align: 'right',
          render: (row) => (
            <span className="font-semibold">{formatCurrency(row.revenue || row.total_revenue || 0)}</span>
          ),
        },
        {
          key: 'average_order_value',
          header: 'Avg Order Value',
          align: 'right',
          render: (row) =>
            formatCurrency(
              (row.revenue || row.total_revenue || 0) /
                Math.max(1, row.orders_count || row.total_orders || 1)
            ),
        },
      ];
    }

    if (activeReport === 'orders') {
      return [
        { key: 'date', header: 'Date' },
        { key: 'total_orders', header: 'Orders Placed', align: 'center' },
        { key: 'completed_orders', header: 'Delivered', align: 'center' },
        { key: 'cancelled_orders', header: 'Cancelled', align: 'center' },
        {
          key: 'total_value',
          header: 'Gross Value',
          align: 'right',
          render: (row) => (
            <span className="font-semibold">{formatCurrency(row.total_value || 0)}</span>
          ),
        },
      ];
    }

    return [
      { key: 'product_name', header: 'Product Name' },
      { key: 'sku', header: 'SKU / Variant' },
      { key: 'units_sold', header: 'Units Sold', align: 'center' },
      {
        key: 'revenue',
        header: 'Total Revenue',
        align: 'right',
        render: (row) => (
          <span className="font-semibold">{formatCurrency(row.revenue || 0)}</span>
        ),
      },
    ];
  };

  const tableData =
    activeReport === 'sales'
      ? reportData?.sales || reportData?.items || []
      : activeReport === 'orders'
      ? reportData?.orders || reportData?.items || []
      : reportData?.products || reportData?.items || [];

  return (
    <div className="admin-reports-page">
      {/* Header */}
      <div className="admin-page-header">
        <div>
          <span className="admin-page-eyebrow">Business Intelligence</span>
          <h1 className="admin-page-title">Executive Reports & Exports</h1>
          <p className="admin-page-description">
            Audit operational velocity, track sales trajectory, and export board-ready analytics.
          </p>
        </div>
        <div className="admin-header-actions">
          <Button
            variant="outline"
            size="sm"
            icon={<FileSpreadsheet size={14} />}
            onClick={() => handleExport('excel')}
            disabled={isExporting}
          >
            Export Excel
          </Button>

          {activeReport !== 'products' && (
            <Button
              variant="outline"
              size="sm"
              icon={<FileText size={14} />}
              onClick={() => handleExport('pdf')}
              disabled={isExporting}
            >
              Export PDF
            </Button>
          )}

          <Button
            variant="secondary"
            size="sm"
            icon={<RefreshCw size={14} className={isLoading ? 'spin-icon' : ''} />}
            onClick={fetchReport}
            disabled={isLoading}
          >
            Refresh
          </Button>
        </div>
      </div>

      {/* Report Selection Tabs */}
      <div className="admin-tabs">
        <button
          type="button"
          className={`admin-tab-btn ${activeReport === 'sales' ? 'active' : ''}`}
          onClick={() => setActiveReport('sales')}
        >
          Sales & Revenue
        </button>
        <button
          type="button"
          className={`admin-tab-btn ${activeReport === 'orders' ? 'active' : ''}`}
          onClick={() => setActiveReport('orders')}
        >
          Order Velocity
        </button>
        <button
          type="button"
          className={`admin-tab-btn ${activeReport === 'products' ? 'active' : ''}`}
          onClick={() => setActiveReport('products')}
        >
          Top Products & Inventory
        </button>
      </div>

      {/* Date & Filter Toolbar */}
      <div className="admin-toolbar">
        <div className="admin-presets-group">
          <Button variant="ghost" size="sm" onClick={() => applyPreset(7)}>
            Last 7 Days
          </Button>
          <Button variant="ghost" size="sm" onClick={() => applyPreset(30)}>
            Last 30 Days
          </Button>
          <Button variant="ghost" size="sm" onClick={() => applyPreset(90)}>
            Last 90 Days
          </Button>
        </div>

        <div className="admin-date-picker-group">
          <div className="admin-date-input-wrapper">
            <span className="admin-date-label">From:</span>
            <input
              type="date"
              value={startDate}
              onChange={(e) => setStartDate(e.target.value)}
              className="admin-date-input"
            />
          </div>

          <div className="admin-date-input-wrapper">
            <span className="admin-date-label">To:</span>
            <input
              type="date"
              value={endDate}
              onChange={(e) => setEndDate(e.target.value)}
              className="admin-date-input"
            />
          </div>

          {activeReport !== 'products' && (
            <select
              value={groupBy}
              onChange={(e) => setGroupBy(e.target.value)}
              className="admin-select"
            >
              <option value="day">Group by Day</option>
              <option value="week">Group by Week</option>
              <option value="month">Group by Month</option>
            </select>
          )}
        </div>
      </div>

      {/* KPI Overview Summary */}
      <div className="admin-report-kpis">
        <div className="admin-kpi-card">
          <span className="admin-kpi-title">Total Gross Revenue</span>
          <span className="admin-kpi-value font-heading">
            {formatCurrency(
              reportData?.total_revenue ||
                reportData?.summary?.total_revenue ||
                reportData?.total_sales ||
                0
            )}
          </span>
        </div>
        <div className="admin-kpi-card">
          <span className="admin-kpi-title">Total Orders Count</span>
          <span className="admin-kpi-value">
            {reportData?.total_orders ||
              reportData?.summary?.total_orders ||
              tableData.length}
          </span>
        </div>
        <div className="admin-kpi-card">
          <span className="admin-kpi-title">Average Order Value</span>
          <span className="admin-kpi-value font-heading">
            {formatCurrency(
              (reportData?.total_revenue || 0) /
                Math.max(1, reportData?.total_orders || 1)
            )}
          </span>
        </div>
      </div>

      {/* Data Table */}
      <DataTable
        columns={getColumns()}
        data={tableData}
        keyField="date"
        isLoading={isLoading}
        emptyMessage="No analytics data available for the specified range."
      />
    </div>
  );
}
