import { api } from '../api';

export type VerificationType = 'REGISTER' | 'FORGOT_PASSWORD' | 'CHANGE_PASSWORD';

export interface AuthSessionResponse {
  message: string;
  access_token: string;
  sid: string;
  access_token_expires_at: string;
}

export interface MessageResponse {
  message: string;
}

export interface RegisterParams {
  name: string;
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  gender: string;
  birthdayDate: string;
  facultyId: number | string;
  avatar?: File | null;
  banner?: File | null;
  bio?: string;
}

export interface LoginParams {
  email: string;
  password: string;
  otp: string;
}

export interface SendLoginOtpParams {
  email: string;
  password: string;
}

export interface ChangePasswordParams {
  newPassword: string;
  confirmPassword: string;
}

export interface ForgotPasswordParams {
  email: string;
  newPassword: string;
  confirmPassword: string;
}

function appendOptionalString(formData: FormData, key: string, value?: string | number): void {
  if (value !== undefined && value !== '') {
    formData.append(key, String(value));
  }
}

export function register(params: RegisterParams): Promise<MessageResponse> {
  const formData = new FormData();
  formData.append('name', params.name);
  formData.append('username', params.username);
  formData.append('email', params.email);
  formData.append('password', params.password);
  formData.append('confirm_password', params.confirmPassword);
  formData.append('gender', params.gender);
  formData.append('birthday_date', params.birthdayDate);
  formData.append('faculty_id', String(params.facultyId));
  appendOptionalString(formData, 'bio', params.bio);

  if (params.avatar) {
    formData.append('avatar', params.avatar);
  }
  if (params.banner) {
    formData.append('banner', params.banner);
  }

  return api.post<MessageResponse>('/api/auth/register', formData);
}

export function sendLoginOtp(params: SendLoginOtpParams): Promise<MessageResponse> {
  return api.post<MessageResponse>('/api/auth/login/otp/send', {
    email: params.email,
    password: params.password,
  });
}

export function login(params: LoginParams): Promise<AuthSessionResponse> {
  return api.post<AuthSessionResponse>('/api/auth/login', {
    email: params.email,
    password: params.password,
    otp: params.otp,
  });
}

export function loginProd(email: string, password: string): Promise<AuthSessionResponse> {
  return api.post<AuthSessionResponse>('/api/auth/login/prod', { email, password });
}

export function refreshAuth(sid: string): Promise<AuthSessionResponse> {
  return api.post<AuthSessionResponse>('/api/auth/refresh', { sid });
}

export function logout(token: string): Promise<MessageResponse> {
  return api.post<MessageResponse>('/api/auth/logout', undefined, token);
}

export function changePasswordWithOtp(
  params: ChangePasswordParams,
  token: string
): Promise<MessageResponse> {
  return api.put<MessageResponse>(
    '/api/auth/change-password',
    {
      new_password: params.newPassword,
      confirm_password: params.confirmPassword,
    },
    token
  );
}

export function forgotPassword(params: ForgotPasswordParams): Promise<MessageResponse> {
  return api.post<MessageResponse>('/api/auth/forgot-password', {
    email: params.email,
    new_password: params.newPassword,
    confirm_password: params.confirmPassword,
  });
}

// Public availability probe for the signup form (no token).
export function checkUsernameAvailable(username: string): Promise<{ available: boolean }> {
  return api.get<{ available: boolean }>(
    `/api/auth/username-available?username=${encodeURIComponent(username)}`
  );
}

export function sendVerification(
  email: string,
  type: VerificationType
): Promise<MessageResponse> {
  return api.post<MessageResponse>('/api/verification/send', { email, type });
}

export function verifyVerification(
  email: string,
  code: string,
  type: VerificationType
): Promise<MessageResponse> {
  return api.post<MessageResponse>('/api/verification/verify', { email, code, type });
}
