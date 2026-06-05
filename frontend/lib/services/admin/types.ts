export interface PaginatedMeta {
  total: number;
  page: number;
  limit: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: PaginatedMeta;
}

export interface MessageResponse {
  message: string;
}

export interface AdminUser {
  user_id: number;
  name: string;
  username: string;
  email: string;
  role_id: number;
  role_name: string;
  created_at?: string;
  avatar_url?: string | null;
  banner_url?: string | null;
  bio?: string;
  is_premium?: boolean;
  gender?: string;
  birthday_date?: string;
  follower_count?: number;
  following_count?: number;
  onboarding_completed?: boolean;
  faculty?: { faculty_id: number; faculty_name: string } | null;
  faculty_location?: { id: number; name: string; region: string } | null;
}

export interface Violation {
  id: string;
  name: string;
  description: string;
  created_at?: string;
  updated_at?: string;
}

export interface Category {
  id: string;
  name: string;
  created_at?: string;
}

export interface Faculty {
  faculty_id: number;
  faculty_name: string;
  total_users?: number;
}

export interface ReportUser {
  user_id: number;
  username: string;
  avatar_url?: string | null;
  is_premium?: boolean;
  faculty?: { faculty_id: number; faculty_name: string } | null;
}

export interface Report {
  id: string;
  reporter_id: number;
  reported_user_id?: number | null;
  reported_post_id?: number | null;
  violation_id: string;
  violation?: Violation;
  reporter_user?: ReportUser;
  reported_user?: ReportUser | null;
  explanation: string;
  status: string;
  created_at?: string;
  updated_at?: string;
}

export interface AdminEventReview {
  revision_number: number;
  notes: string;
  admin_name?: string;
  created_at: string;
}

// AdminEventAudience mirrors the frontend EventAudience shape (camelCase) so the
// shared <AudienceBadges /> component can render admin event region chips.
export interface AdminEventAudience {
  id: number;
  name: string;
  region: string;
  isAllRegions: boolean;
}

export interface AdminEvent {
  id: string;
  user_id: number;
  title: string;
  description: string;
  location: string;
  type: string;
  status: string;
  start_time: string;
  end_time: string;
  banner_url?: string | null;
  attachment_url?: string | null;
  doc_urls?: string[];
  audience?: AdminEventAudience[];
  created_at?: string;
  updated_at?: string;
  notes?: AdminEventReview[];
}

export interface TrendReasonStats {
  code: string;
  label: string;
  count: number;
}

export interface TrendReason {
  code: string;
  label: string;
}

export interface Order {
  id: number;
  order_id: string;
  user_id: number;
  amount: number;
  purpose: string;
  status: string;
  payment_type?: string;
  created_at?: string;
  updated_at?: string;
}

export interface FeedbackAuthor {
  user_id: number;
  username: string;
  name?: string;
  avatar?: string | null;
  user_avatar?: string | null;
  is_premium?: boolean;
}

export interface Feedback {
  id: string;
  type: string;
  title: string;
  body: string;
  status: string;
  images?: string[];
  author?: FeedbackAuthor;
  likes?: number;
  dislikes?: number;
  created_at?: string;
  updated_at?: string;
}

export interface Role {
  role_id: number;
  name: string;
}

export interface SchemaColumn {
  name: string;
  type: string;
  nullable: boolean;
  is_pk: boolean;
  is_unique: boolean;
  is_fk: boolean;
  fk_table?: string;
  fk_column?: string;
}

export interface SchemaReference {
  from_table: string;
  from_column: string;
  to_column: string;
}

export interface SchemaTable {
  name: string;
  columns: SchemaColumn[];
  referenced_by: SchemaReference[];
}

export interface SchemaResponse {
  tables: SchemaTable[];
}
