export class ApiError extends Error {
  constructor({ message, status = 0, data = null, isNetworkError = false }) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.data = data;
    this.isNetworkError = isNetworkError;
  }
}

export function toApiError(error) {
  if (error instanceof ApiError) {
    return error;
  }

  const response = error?.response;
  const data = response?.data ?? null;

  return new ApiError({
    message:
      data?.message ||
      (typeof data?.error === 'string' ? data?.error : null) ||
      error?.message ||
      'Unable to complete the request.',
    status: response?.status ?? 0,
    data,
    isNetworkError: !response,
  });
}
