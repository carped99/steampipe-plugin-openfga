-- Steampipe ACL FDW Initialization Script
-- This script sets up the Steampipe Postgres ACL extension

-- Create steampipe_postgres_acl extension
CREATE EXTENSION IF NOT EXISTS steampipe_postgres_acl;

-- The extension automatically creates the acl schema and tables

-- Grant permissions
GRANT USAGE ON SCHEMA acl TO PUBLIC;
GRANT SELECT ON ALL TABLES IN SCHEMA acl TO PUBLIC;

COMMENT ON SCHEMA acl IS 'OpenFGA Access Control List integration via Steampipe';

-- Example: Create a helper function for permission checks
CREATE OR REPLACE FUNCTION acl.check_permission(
  p_subject_type TEXT,
  p_subject_id TEXT,
  p_relation TEXT,
  p_object_type TEXT,
  p_object_id TEXT
)
RETURNS BOOLEAN AS $$
  SELECT allowed
  FROM acl.sys_acl_permission
  WHERE subject_type = p_subject_type
    AND subject_id = p_subject_id
    AND relation = p_relation
    AND object_type = p_object_type
    AND object_id = p_object_id;
$$ LANGUAGE SQL STABLE;

COMMENT ON FUNCTION acl.check_permission IS 'Check if a subject has permission on an object';

-- Log initialization
DO $$
BEGIN
  RAISE NOTICE 'Steampipe ACL plugin initialized successfully';
  RAISE NOTICE 'Available tables:';
  RAISE NOTICE '  - acl.sys_acl_permission';
  RAISE NOTICE 'Helper functions:';
  RAISE NOTICE '  - acl.check_permission(subject_type, subject_id, relation, object_type, object_id)';
END $$;
