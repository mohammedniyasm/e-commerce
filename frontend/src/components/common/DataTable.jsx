import { ChevronLeft, ChevronRight, ArrowUpDown, Inbox } from 'lucide-react';
import { SkeletonLoader } from './SkeletonLoader';
import { Button } from './Button';
import './DataTable.css';

/**
 * Drape Reusable Admin Data Table
 * Provides sorting, pagination, skeleton loading, empty state, and responsive layout
 */
export function DataTable({
  columns = [],
  data = [],
  keyField = 'id',
  isLoading = false,
  emptyMessage = 'No records found.',
  emptyAction = null,
  sortColumn = null,
  sortDirection = 'asc', // 'asc' | 'desc'
  onSort = null,
  onRowClick = null,
  pagination = null, // { page, totalPages, totalItems, limit, onPageChange }
  className = '',
}) {
  const handleSort = (columnKey, sortable) => {
    if (!sortable || !onSort) return;
    const nextDirection = sortColumn === columnKey && sortDirection === 'asc' ? 'desc' : 'asc';
    onSort(columnKey, nextDirection);
  };

  return (
    <div className={`drape-table-card ${className}`}>
      <div className="drape-table-container">
        <table className="drape-table">
          <thead>
            <tr>
              {columns.map((col) => {
                const isSorted = sortColumn === col.key;
                return (
                  <th
                    key={col.key}
                    style={{ width: col.width, textAlign: col.align || 'left' }}
                    className={col.sortable ? 'drape-th--sortable' : ''}
                    onClick={() => handleSort(col.key, col.sortable)}
                    aria-sort={isSorted ? (sortDirection === 'asc' ? 'ascending' : 'descending') : undefined}
                  >
                    <div className="drape-th__content" style={{ justifyContent: col.align === 'right' ? 'flex-end' : col.align === 'center' ? 'center' : 'flex-start' }}>
                      <span>{col.header}</span>
                      {col.sortable && (
                        <ArrowUpDown
                          size={13}
                          className={`drape-th__sort-icon ${isSorted ? 'drape-th__sort-icon--active' : ''}`}
                        />
                      )}
                    </div>
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {isLoading ? (
              Array.from({ length: 5 }).map((_, rowIndex) => (
                <tr key={`skeleton-${rowIndex}`} className="drape-tr--skeleton">
                  {columns.map((col) => (
                    <td key={col.key}>
                      <SkeletonLoader height="16px" width="80%" />
                    </td>
                  ))}
                </tr>
              ))
            ) : data.length === 0 ? (
              <tr className="drape-tr--empty">
                <td colSpan={columns.length}>
                  <div className="drape-table-empty">
                    <Inbox size={40} className="drape-table-empty__icon" aria-hidden="true" />
                    <p className="drape-table-empty__title">No items found</p>
                    <p className="drape-table-empty__desc">{emptyMessage}</p>
                    {emptyAction && <div className="drape-table-empty__action">{emptyAction}</div>}
                  </div>
                </td>
              </tr>
            ) : (
              data.map((row, index) => (
                <tr
                  key={row[keyField] ?? index}
                  className={onRowClick ? 'drape-tr--clickable' : ''}
                  onClick={onRowClick ? () => onRowClick(row) : undefined}
                >
                  {columns.map((col) => {
                    const cellValue = row[col.key];
                    let renderedContent = '—';
                    if (col.render) {
                      try {
                        renderedContent = col.render(cellValue !== undefined ? cellValue : row, row);
                      } catch (_e1) {
                        try {
                          renderedContent = col.render(row, row);
                        } catch (err) {
                          console.error(`Error rendering column ${col.key}:`, err);
                          renderedContent = cellValue ?? '—';
                        }
                      }
                    } else {
                      renderedContent = cellValue ?? '—';
                    }

                    return (
                      <td key={col.key} style={{ textAlign: col.align || 'left' }}>
                        {renderedContent}
                      </td>
                    );
                  })}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination Footer */}
      {pagination && pagination.totalPages > 1 && (
        <div className="drape-table-pagination">
          <div className="drape-pagination-info">
            Showing{' '}
            <span className="drape-pagination-info__number">
              {(pagination.page - 1) * pagination.limit + 1}
            </span>{' '}
            to{' '}
            <span className="drape-pagination-info__number">
              {Math.min(pagination.page * pagination.limit, pagination.totalItems || 0)}
            </span>{' '}
            of{' '}
            <span className="drape-pagination-info__number">{pagination.totalItems}</span> items
          </div>

          <div className="drape-pagination-controls">
            <Button
              variant="outline"
              size="sm"
              disabled={pagination.page <= 1 || isLoading}
              onClick={() => pagination.onPageChange(pagination.page - 1)}
              aria-label="Previous page"
              leftIcon={<ChevronLeft size={16} />}
            >
              Prev
            </Button>

            <span className="drape-pagination-page-indicator">
              Page {pagination.page} of {pagination.totalPages}
            </span>

            <Button
              variant="outline"
              size="sm"
              disabled={pagination.page >= pagination.totalPages || isLoading}
              onClick={() => pagination.onPageChange(pagination.page + 1)}
              aria-label="Next page"
              rightIcon={<ChevronRight size={16} />}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
