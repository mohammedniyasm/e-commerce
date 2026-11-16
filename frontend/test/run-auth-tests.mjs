import { createServer } from 'vite';
import React from 'react';
import { renderToString } from 'react-dom/server';
import { MemoryRouter } from 'react-router-dom';

async function run() {
  console.log('--- STARTING CUSTOMER AUTHENTICATION VERIFICATION TESTS ---');

  const vite = await createServer({
    server: { middlewareMode: true },
    appType: 'custom',
    root: process.cwd(),
  });

  try {
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

    // 1. Test Validators
    console.log('\n[1/3] Testing Form Validators...');
    const Validators = await vite.ssrLoadModule('/src/utils/validators.js');
    const {
      validateName,
      cleanPhoneNumber,
      validatePhone,
      validateEmail,
      checkPasswordRules,
      validatePassword,
      validateConfirmPassword,
      validateOTP,
    } = Validators;

    // Name validation
    assert(validateName('Eleni Vance') === null, 'validateName accepts valid name');
    assert(validateName('A') !== null, 'validateName rejects single character name');
    assert(validateName('Name With Numbers 123') !== null, 'validateName rejects numbers');
    assert(validateName('Jean-Luc D\'Souza') === null, 'validateName accepts hyphens and apostrophes');

    // Phone validation
    assert(cleanPhoneNumber('+91 98765 43210') === '9876543210', 'cleanPhoneNumber cleans prefix and whitespace');
    assert(validatePhone('9876543210') === null, 'validatePhone accepts 10-digit phone');
    assert(validatePhone('+91 98765 43210') === null, 'validatePhone accepts formatted phone with +91');
    assert(validatePhone('12345') !== null, 'validatePhone rejects short phone');
    assert(validatePhone('98765abcde') !== null, 'validatePhone rejects alphabetic phone');

    // Email validation
    assert(validateEmail('name@atelier-drape.com') === null, 'validateEmail accepts standard email');
    assert(validateEmail('invalid-email') !== null, 'validateEmail rejects email without @');
    assert(validateEmail('test@') !== null, 'validateEmail rejects missing domain');

    // Password validation (min 8 chars, 1 upper, 1 lower, 1 digit, 1 special)
    const weakCheck = checkPasswordRules('short');
    assert(!weakCheck.isValid && !weakCheck.minLength, 'checkPasswordRules identifies short password');

    const strongCheck = checkPasswordRules('StrongP@ss1');
    assert(
      strongCheck.isValid &&
        strongCheck.minLength &&
        strongCheck.hasUppercase &&
        strongCheck.hasLowercase &&
        strongCheck.hasNumber &&
        strongCheck.hasSpecial,
      'checkPasswordRules validates strong password satisfying all criteria',
    );

    assert(validatePassword('StrongP@ss1') === null, 'validatePassword accepts strong password');
    assert(validatePassword('nouppercase1@') !== null, 'validatePassword rejects missing uppercase');
    assert(validatePassword('NOLOWERCASE1@') !== null, 'validatePassword rejects missing lowercase');
    assert(validatePassword('NoNumberSpecial!') !== null, 'validatePassword rejects missing digit');
    assert(validatePassword('NoSpecial1234') !== null, 'validatePassword rejects missing special char');

    // Confirm Password
    assert(validateConfirmPassword('Pass123!', 'Pass123!') === null, 'validateConfirmPassword accepts matching passwords');
    assert(validateConfirmPassword('Pass123!', 'Mismatch!') !== null, 'validateConfirmPassword rejects mismatched passwords');

    // OTP validation
    assert(validateOTP('123456') === null, 'validateOTP accepts 6 digits');
    assert(validateOTP('12345') !== null, 'validateOTP rejects 5 digits');
    assert(validateOTP('12345a') !== null, 'validateOTP rejects non-digits');

    // 2. Test API Service Contract Exports
    console.log('\n[2/3] Testing API Service Functions...');
    const AuthService = await vite.ssrLoadModule('/src/api/authService.js');
    assert(typeof AuthService.registerCustomer === 'function', 'authService exports registerCustomer');
    assert(typeof AuthService.loginCustomer === 'function', 'authService exports loginCustomer');
    assert(typeof AuthService.loginWithGoogle === 'function', 'authService exports loginWithGoogle');
    assert(typeof AuthService.forgotPassword === 'function', 'authService exports forgotPassword');
    assert(typeof AuthService.verifyForgotPasswordOTP === 'function', 'authService exports verifyForgotPasswordOTP');
    assert(typeof AuthService.resendForgotPasswordOTP === 'function', 'authService exports resendForgotPasswordOTP');
    assert(typeof AuthService.resetPassword === 'function', 'authService exports resetPassword');

    // 3. Test Component Rendering
    console.log('\n[3/3] Testing Component SSR Rendering...');
    const { AuthProvider } = await vite.ssrLoadModule('/src/context/AuthContext.jsx');
    const { ToastProvider } = await vite.ssrLoadModule('/src/context/ToastContext.jsx');

    function renderWithProviders(element) {
      return renderToString(
        React.createElement(
          MemoryRouter,
          null,
          React.createElement(
            ToastProvider,
            null,
            React.createElement(AuthProvider, null, element),
          ),
        ),
      );
    }

    // RegisterPage
    const { RegisterPage } = await vite.ssrLoadModule('/src/pages/auth/RegisterPage.jsx');
    const registerHtml = renderWithProviders(React.createElement(RegisterPage));
    assert(registerHtml.includes('Create Account'), 'RegisterPage renders title "Create Account"');
    assert(registerHtml.includes('FULL NAME'), 'RegisterPage renders FULL NAME label');
    assert(registerHtml.includes('EMAIL ADDRESS'), 'RegisterPage renders EMAIL ADDRESS label');
    assert(registerHtml.includes('PHONE NUMBER'), 'RegisterPage renders PHONE NUMBER label');
    assert(registerHtml.includes('+91'), 'RegisterPage renders +91 prefix badge');
    assert(registerHtml.includes('PASSWORD'), 'RegisterPage renders PASSWORD label');
    assert(registerHtml.includes('CREATE ACCOUNT'), 'RegisterPage renders CREATE ACCOUNT submit button');
    assert(registerHtml.includes('Continue with Google'), 'RegisterPage renders Continue with Google button');
    assert(registerHtml.includes('Already have an account?'), 'RegisterPage renders login redirect link');

    // LoginPage
    const { LoginPage } = await vite.ssrLoadModule('/src/pages/auth/LoginPage.jsx');
    const loginHtml = renderWithProviders(React.createElement(LoginPage));
    assert(loginHtml.includes('Welcome Back'), 'LoginPage renders title "Welcome Back"');
    assert(loginHtml.includes('EMAIL ADDRESS'), 'LoginPage renders EMAIL ADDRESS label');
    assert(loginHtml.includes('PASSWORD'), 'LoginPage renders PASSWORD label');
    assert(loginHtml.includes('Forgot Password?'), 'LoginPage renders Forgot Password link');
    assert(loginHtml.includes('LOGIN'), 'LoginPage renders LOGIN button');
    assert(loginHtml.includes('Create Account'), 'LoginPage renders Create Account redirect link');

    // ForgotPasswordPage
    const { ForgotPasswordPage } = await vite.ssrLoadModule('/src/pages/auth/ForgotPasswordPage.jsx');
    const forgotHtml = renderWithProviders(React.createElement(ForgotPasswordPage));
    assert(forgotHtml.includes('Forgot Password'), 'ForgotPasswordPage renders title "Forgot Password"');
    assert(forgotHtml.includes('SEND OTP'), 'ForgotPasswordPage renders SEND OTP button');
    assert(forgotHtml.includes('Back to Login'), 'ForgotPasswordPage renders Back to Login link');

    // ResetPasswordPage
    const { ResetPasswordPage } = await vite.ssrLoadModule('/src/pages/auth/ResetPasswordPage.jsx');
    const resetHtml = renderWithProviders(React.createElement(ResetPasswordPage));
    assert(resetHtml.includes('Reset Password'), 'ResetPasswordPage renders title "Reset Password"');
    assert(resetHtml.includes('NEW PASSWORD'), 'ResetPasswordPage renders NEW PASSWORD label');
    assert(resetHtml.includes('CONFIRM NEW PASSWORD'), 'ResetPasswordPage renders CONFIRM NEW PASSWORD label');
    assert(resetHtml.includes('RESET PASSWORD'), 'ResetPasswordPage renders RESET PASSWORD submit button');

    // PasswordResetSuccessPage
    const { PasswordResetSuccessPage } = await vite.ssrLoadModule('/src/pages/auth/PasswordResetSuccessPage.jsx');
    const successHtml = renderWithProviders(React.createElement(PasswordResetSuccessPage));
    assert(successHtml.includes('Password Reset Successful'), 'PasswordResetSuccessPage renders success title');
    assert(successHtml.includes('BACK TO LOGIN'), 'PasswordResetSuccessPage renders BACK TO LOGIN button');

    // AuthLayout
    const { AuthLayout } = await vite.ssrLoadModule('/src/layouts/AuthLayout.jsx');
    const layoutHtml = renderWithProviders(React.createElement(AuthLayout));
    assert(layoutHtml.includes('EDITION 01 / SS25'), 'AuthLayout renders editorial edition badge');
    assert(layoutHtml.includes('KOCHI FLAGSHIP'), 'AuthLayout renders flagship location label');
    assert(layoutHtml.includes('Tactile permanence and effortless daily utility.'), 'AuthLayout renders campaign quote');
    assert(layoutHtml.includes('LOOK 14: RELAXED LINEN SUITING'), 'AuthLayout renders editorial look metadata');
    assert(layoutHtml.includes('DRAPE'), 'AuthLayout renders DRAPE brand logo');

    // Summary
    const failed = assertions.filter((a) => !a.pass);
    console.log(`\n--- AUTH TEST SUMMARY: ${assertions.length - failed.length}/${assertions.length} PASSED ---`);

    if (failed.length > 0) {
      console.error(`Failed ${failed.length} test assertions.`);
      process.exit(1);
    } else {
      console.log('ALL CUSTOMER AUTHENTICATION TESTS PASSED SUCCESSFULLY (Exit code 0).');
      process.exit(0);
    }
  } catch (error) {
    console.error('Test Execution Error:', error);
    process.exit(1);
  } finally {
    await vite.close();
  }
}

run();
