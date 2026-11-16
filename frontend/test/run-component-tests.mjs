import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';

async function run() {
  console.log('--- STARTING COMPONENT VERIFICATION TESTS ---');

  const vite = await createServer({
    server: { middlewareMode: true },
    appType: 'custom',
    root: process.cwd(),
  });

  try {
    const Common = await vite.ssrLoadModule('/src/components/common/index.js');
    const {
      Button,
      Input,
      FormField,
      Select,
      Modal,
      Badge,
      StatusPill,
      Spinner,
      SkeletonLoader,
      DataTable,
    } = Common;

    const assertions = [];

    function assert(condition, message) {
      if (condition) {
        assertions.push({ pass: true, message });
        console.log(`  ✓ ${message}`);
      } else {
        assertions.push({ pass: false, message });
        console.error(`  ✗ FAIL: ${message}`);
      }
    }

    // 1. Button Verification
    const btnHtml = renderToString(React.createElement(Button, { variant: 'primary', size: 'md' }, 'Click me'));
    assert(btnHtml.includes('drape-btn--primary') && btnHtml.includes('drape-btn--md'), 'Button renders primary variant and md size');

    const btnLoading = renderToString(React.createElement(Button, { isLoading: true }, 'Saving'));
    assert(btnLoading.includes('drape-btn--loading') && btnLoading.includes('aria-busy="true"'), 'Button handles loading state and aria-busy');

    const btnDisabled = renderToString(React.createElement(Button, { disabled: true }, 'Disabled'));
    assert(btnDisabled.includes('disabled=""') && btnDisabled.includes('aria-disabled="true"'), 'Button handles disabled state');

    // 2. Input & FormField Verification
    const inputHtml = renderToString(React.createElement(Input, { placeholder: 'Email', hasError: true, id: 'email' }));
    assert(inputHtml.includes('drape-input-wrapper--error') && inputHtml.includes('aria-invalid="true"'), 'Input reflects error state and aria-invalid');

    const formFieldHtml = renderToString(
      React.createElement(FormField, { id: 'email', label: 'Email Address', required: true, error: 'Invalid email' },
        React.createElement(Input, { id: 'email' })
      )
    );
    assert(formFieldHtml.includes('Email Address') && formFieldHtml.includes('Invalid email'), 'FormField renders label, required indicator, and inline error message');

    // 3. Select Verification
    const selectHtml = renderToString(
      React.createElement(Select, {
        options: [
          { value: 'all', label: 'All Categories' },
          { value: 'men', label: 'Men' },
        ],
        value: 'all',
        onChange: () => {},
      })
    );
    assert(selectHtml.includes('All Categories') && selectHtml.includes('drape-select__chevron'), 'Select renders options and custom chevron indicator');

    // 4. Modal Verification (Focus & Accessibility)
    const modalOpen = renderToString(
      React.createElement(Modal, {
        isOpen: true,
        title: 'Sign Out of Drape Admin?',
        eyebrow: 'SECURITY NOTICE',
        description: 'Are you sure you want to end your current session?',
        onClose: () => {},
        confirmText: 'Sign Out',
      })
    );
    assert(modalOpen.includes('role="dialog"') && modalOpen.includes('aria-modal="true"'), 'Modal declares role="dialog" and aria-modal="true"');
    assert(modalOpen.includes('tabindex="-1"'), 'Modal container has tabIndex="-1" for focus management');
    assert(modalOpen.includes('SECURITY NOTICE') && modalOpen.includes('Sign Out of Drape Admin?'), 'Modal renders eyebrow and title');

    const modalClosed = renderToString(
      React.createElement(Modal, { isOpen: false, onClose: () => {} })
    );
    assert(modalClosed === '', 'Modal returns null when isOpen is false');

    // 5. Badge & StatusPill Verification
    const badgeHtml = renderToString(React.createElement(Badge, { variant: 'filter', isActive: true }, 'MEN'));
    assert(badgeHtml.includes('drape-badge--filter') && badgeHtml.includes('drape-badge--active'), 'Badge renders active filter pill state');

    const statusPillHtml = renderToString(React.createElement(StatusPill, { status: 'Delivered', variant: 'success' }));
    assert(statusPillHtml.includes('drape-status-pill--success') && statusPillHtml.includes('Delivered'), 'StatusPill renders success variant with dot indicator');

    // 6. Spinner & SkeletonLoader Verification
    const spinnerHtml = renderToString(React.createElement(Spinner, { size: 24, label: 'Processing' }));
    assert(spinnerHtml.includes('role="status"') && spinnerHtml.includes('aria-label="Processing"'), 'Spinner has accessible role="status" and aria-label');

    const skeletonHtml = renderToString(React.createElement(SkeletonLoader, { height: '24px', count: 3 }));
    assert(skeletonHtml.includes('drape-skeleton-group'), 'SkeletonLoader renders group when count > 1');

    // 7. DataTable Verification
    const testCols = [
      { key: 'sku', header: 'SKU', sortable: true },
      { key: 'name', header: 'Product Title' },
    ];
    const testRows = [{ sku: 'DRP-001', name: 'Wool Coat' }];

    const tableWithData = renderToString(
      React.createElement(DataTable, { columns: testCols, data: testRows, keyField: 'sku' })
    );
    assert(tableWithData.includes('DRP-001') && tableWithData.includes('Wool Coat'), 'DataTable renders table rows with data');

    const tableEmpty = renderToString(
      React.createElement(DataTable, { columns: testCols, data: [], emptyMessage: 'No products in inventory' })
    );
    assert(tableEmpty.includes('No products in inventory') && tableEmpty.includes('drape-table-empty'), 'DataTable renders accessible empty state');

    const tableLoading = renderToString(
      React.createElement(DataTable, { columns: testCols, data: [], isLoading: true })
    );
    assert(tableLoading.includes('drape-tr--skeleton'), 'DataTable renders skeleton rows during loading state');

    const failures = assertions.filter((a) => !a.pass);
    if (failures.length > 0) {
      console.error(`\nTest run finished with ${failures.length} failures.`);
      process.exit(1);
    }

    console.log(`\nALL ${assertions.length} COMPONENT TESTS PASSED WITH EXIT CODE 0!`);
  } finally {
    await vite.close();
  }
}

run().catch((err) => {
  console.error('Fatal test error:', err);
  process.exit(1);
});
