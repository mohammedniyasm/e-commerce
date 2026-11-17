import { apiClient } from './client';

export async function getProfile() {
  const response = await apiClient.get('/user/profile', {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function updateProfile({ name, phone }) {
  const response = await apiClient.put(
    '/user/profile',
    { name, phone },
    { requiresAuth: true },
  );
  return response.data?.data || response.data;
}

export async function changePassword({ current_password, new_password }) {
  const response = await apiClient.put(
    '/user/profile/password',
    {
      current_password,
      new_password,
    },
    { requiresAuth: true },
  );
  return response.data?.data || response.data;
}

export async function sendEmailChangeOTP({ new_email }) {
  const response = await apiClient.post(
    '/user/profile/email',
    { new_email },
    { requiresAuth: true },
  );
  return response.data?.data || response.data;
}

export async function verifyEmailChangeOTP({ otp }) {
  const response = await apiClient.post(
    '/user/profile/email/verify',
    { otp },
    { requiresAuth: true },
  );
  return response.data?.data || response.data;
}

export async function uploadProfileImage(file) {
  const formData = new FormData();
  formData.append('image', file);
  const response = await apiClient.post('/user/profile/image', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export async function deleteProfileImage() {
  const response = await apiClient.delete('/user/profile/image', {
    requiresAuth: true,
  });
  return response.data?.data || response.data;
}

export {
  sendVerificationOTP,
  verifyEmailOTP,
  resendVerificationOTP,
} from './authService';
