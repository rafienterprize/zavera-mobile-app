-- ============================================
-- FIX ADMIN AUDIT LOG - ALLOW NULL admin_user_id
-- ============================================
-- This migration allows admin_user_id to be NULL
-- for cases where admin context is not available
-- (e.g., system-triggered actions, API calls without user context)
-- ============================================

-- Make admin_user_id nullable
ALTER TABLE admin_audit_log 
ALTER COLUMN admin_user_id DROP NOT NULL;

-- Add comment
COMMENT ON COLUMN admin_audit_log.admin_user_id IS 'Admin user ID - nullable for system actions or when user context unavailable';

SELECT 'Admin audit log fixed - admin_user_id now nullable' AS status;
