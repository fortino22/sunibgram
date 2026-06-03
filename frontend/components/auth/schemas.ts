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
    password: z
      .string()
      .min(10, 'Password must be at least 10 characters')
      .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
      .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
      .regex(/[0-9]/, 'Password must contain at least one number'),
    confirmPassword: z.string(),
    gender: z.enum(['male', 'female', 'other', 'prefer_not_to_say'], 'Gender is required'),
    birthdayDate: z
      .string()
      .optional()
      .refine((val) => {
        if (!val) return true;
        const d = new Date(val);
        if (isNaN(d.getTime())) return false;
        const cutoff = new Date();
        cutoff.setFullYear(cutoff.getFullYear() - 17);
        return d <= cutoff;
      }, 'You must be at least 17 years old'),
    facultyId: z.string().min(1, 'Faculty ID is required'),
    facultyLocationId: z.string().min(1, 'Campus is required'),
    bio: z.string().max(500, 'Bio too long').optional(),
    agreeToTerms: z.literal(true, { message: 'You must accept the Terms and Conditions' }),
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
  password: z.string().min(1, 'Password is required'),
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
    newPassword: z
      .string()
      .min(10, 'Password must be at least 10 characters')
      .regex(/[A-Z]/, 'Password must contain at least one uppercase letter')
      .regex(/[a-z]/, 'Password must contain at least one lowercase letter')
      .regex(/[0-9]/, 'Password must contain at least one number'),
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
export type SignupStep = 1 | 2 | 3 | 4;
export type ForgotPasswordStep = 'email' | 'otp' | 'reset' | 'done';
