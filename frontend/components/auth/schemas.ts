import { z } from 'zod';

export const signupSchema = z
  .object({
    name: z.string().min(2, 'Name must be at least 2 characters'),
    username: z
      .string()
      .min(3, 'Username must be at least 3 characters')
      .max(30, 'Username too long')
      .regex(/^[a-zA-Z0-9_]+$/, 'Only letters, numbers, underscores'),
    email: z
      .string()
      .email('Invalid email address')
      .endsWith('@binus.ac.id', 'Must use your @binus.ac.id email'),
    password: z.string().min(8, 'Password must be at least 8 characters'),
    confirmPassword: z.string(),
  })
  .refine((data) => data.password === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

export const loginSchema = z.object({
  email: z
    .string()
    .email('Invalid email address')
    .endsWith('@binus.ac.id', 'Must use your @binus.ac.id email'),
  password: z.string().min(6, 'Password must be at least 6 characters'),
});

export const verificationCodeSchema = z.object({
  code: z
    .string()
    .regex(/^\d{6}$/, 'Enter the 6-digit code'),
});

export const forgotPasswordEmailSchema = z.object({
  email: z
    .string()
    .email('Invalid email address')
    .endsWith('@binus.ac.id', 'Must use your @binus.ac.id email'),
});

export const forgotPasswordResetSchema = z
  .object({
    newPassword: z.string().min(6, 'New password must be at least 6 characters'),
    confirmPassword: z.string(),
  })
  .refine((data) => data.newPassword === data.confirmPassword, {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  });

export type SignupInput = z.infer<typeof signupSchema>;
export type LoginInput = z.infer<typeof loginSchema>;
export type VerificationCodeInput = z.infer<typeof verificationCodeSchema>;
export type ForgotPasswordEmailInput = z.infer<typeof forgotPasswordEmailSchema>;
export type ForgotPasswordResetInput = z.infer<typeof forgotPasswordResetSchema>;
export type SignupStep = 1 | 2 | 3;
export type LoginStep = 'credentials' | 'otp';
export type ForgotPasswordStep = 'email' | 'otp' | 'reset' | 'done';
