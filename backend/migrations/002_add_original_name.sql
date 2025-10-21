-- Add original_name column to files table for existing databases
ALTER TABLE files ADD COLUMN IF NOT EXISTS original_name VARCHAR(255);

-- Update existing records to set original_name to filename
UPDATE files SET original_name = filename WHERE original_name IS NULL;

-- Make original_name NOT NULL after updating existing records
ALTER TABLE files ALTER COLUMN original_name SET NOT NULL;