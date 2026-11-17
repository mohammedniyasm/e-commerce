import assert from 'node:assert/strict';
import { createServer } from 'vite';

console.log('====================================================');
console.log('CUSTOMER EMAIL VERIFICATION VERIFICATION TEST SUITE');
console.log('====================================================\n');

async function runTests() {
  const vite = await createServer({
    server: { middlewareMode: true },
    appType: 'custom',
    root: process.cwd(),
  });

  try {
    console.log('[1/4] Verifying authService API verification exports...');
    const AuthService = await vite.ssrLoadModule('/src/api/authService.js');
    assert(typeof AuthService.sendVerificationOTP === 'function', 'authService exports sendVerificationOTP');
    assert(typeof AuthService.verifyEmailOTP === 'function', 'authService exports verifyEmailOTP');
    assert(typeof AuthService.resendVerificationOTP === 'function', 'authService exports resendVerificationOTP');
    console.log('  ✓ sendVerificationOTP, verifyEmailOTP, and resendVerificationOTP are properly exported\n');

    console.log('[2/4] Verifying EmailVerificationModal component module evaluation...');
    const { EmailVerificationModal } = await vite.ssrLoadModule('/src/components/account/EmailVerificationModal.jsx');
    assert(typeof EmailVerificationModal === 'function', 'EmailVerificationModal is a valid React component');
    console.log('  ✓ EmailVerificationModal compiles and exports cleanly\n');

    console.log('[3/4] Verifying UnverifiedAccountBanner & UnverifiedAccountPopup components...');
    const { UnverifiedAccountBanner, UnverifiedAccountPopup } = await vite.ssrLoadModule('/src/components/account/UnverifiedAccountBanner.jsx');
    assert(typeof UnverifiedAccountBanner === 'function', 'UnverifiedAccountBanner is a valid React component');
    assert(typeof UnverifiedAccountPopup === 'function', 'UnverifiedAccountPopup is a valid React component');
    console.log('  ✓ UnverifiedAccountBanner and UnverifiedAccountPopup evaluate cleanly\n');

    console.log('[4/4] Verifying ProfilePage & AccountOverviewPage integration...');
    const { ProfilePage } = await vite.ssrLoadModule('/src/pages/account/ProfilePage.jsx');
    const { AccountOverviewPage } = await vite.ssrLoadModule('/src/pages/account/AccountOverviewPage.jsx');
    const { AccountLayout } = await vite.ssrLoadModule('/src/layouts/AccountLayout.jsx');
    assert(typeof ProfilePage === 'function', 'ProfilePage is valid');
    assert(typeof AccountOverviewPage === 'function', 'AccountOverviewPage is valid');
    assert(typeof AccountLayout === 'function', 'AccountLayout is valid');
    console.log('  ✓ AccountLayout, ProfilePage, and AccountOverviewPage import and evaluate cleanly\n');

    console.log('====================================================');
    console.log('EMAIL VERIFICATION SUITE: ALL TESTS PASSED (0 FAILED)');
    console.log('====================================================');
  } finally {
    await vite.close();
  }
}

runTests().catch((err) => {
  console.error('Test Suite Failed:', err);
  process.exit(1);
});
