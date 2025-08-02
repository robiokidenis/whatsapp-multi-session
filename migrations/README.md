# Database Migrations

This directory contains database migration scripts for the WhatsApp Multi-Session API.

## Auto Reply Text Feature Migration

The `auto_reply_text` column was added to the `session_metadata` table to support session-level auto reply functionality.

### For SQLite (Default)

If you're using SQLite (default configuration), the migration has been automatically applied to your existing database.

To manually apply the migration to a fresh database or verify it exists:

```bash
sqlite3 database/session_metadata.db < migrations/add_auto_reply_text_column.sql
```

### For MySQL

If you're using MySQL, you have two options:

#### Option 1: For New Installations
New MySQL installations will automatically include the `auto_reply_text` column as it's now part of the initialization script (`mysql-init/01-init-database.sql`).

#### Option 2: For Existing MySQL Databases
If you have an existing MySQL database that needs to be migrated:

```bash
# Connect to your MySQL server and run the migration
mysql -u your_username -p whatsapp_multi_session < migrations/mysql_add_auto_reply_text.sql
```

Or manually execute:

```sql
USE whatsapp_multi_session;
ALTER TABLE session_metadata ADD COLUMN auto_reply_text TEXT AFTER webhook_url;
```

### Verification

To verify the migration was successful, check the table structure:

#### SQLite:
```bash
sqlite3 database/session_metadata.db ".schema session_metadata"
```

#### MySQL:
```sql
USE whatsapp_multi_session;
DESCRIBE session_metadata;
```

You should see the `auto_reply_text TEXT` column in the table definition.

### Rollback

If you need to remove the `auto_reply_text` column:

#### SQLite:
```sql
-- SQLite doesn't support DROP COLUMN directly
-- You would need to recreate the table without the column
-- This is generally not recommended for production
```

#### MySQL:
```sql
USE whatsapp_multi_session;
ALTER TABLE session_metadata DROP COLUMN auto_reply_text;
```

**Note:** Rolling back this migration will cause the auto reply feature to stop working.

## Migration History

- `2025-08-02`: Added `auto_reply_text` column to `session_metadata` table
  - Files: `add_auto_reply_text_column.sql`, `mysql_add_auto_reply_text.sql`
  - Purpose: Enable session-level auto reply functionality
  - Type: `TEXT` (nullable)