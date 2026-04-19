# Database Migrations

## Running Migrations

### Apply All Migrations

```bash
# Using Makefile
make migrate-up

# Or directly with psql
psql $DATABASE_URL -f migrations/001_initial_schema.up.sql
psql $DATABASE_URL -f migrations/002_add_voice_settings.up.sql
```

### Rollback Migrations

```bash
# Using Makefile
make migrate-down

# Or directly with psql (in reverse order)
psql $DATABASE_URL -f migrations/002_add_voice_settings.down.sql
psql $DATABASE_URL -f migrations/001_initial_schema.down.sql
```

## Migration List

### 001_initial_schema
**Created**: Initial setup
**Description**: Creates all core tables (projects, scripts, scenes, audio_files, assets, subtitles, renders, jobs, review_actions)

### 002_add_voice_settings
**Created**: 2024
**Description**: Adds voice selection fields to projects table
- `voice_id VARCHAR(100) DEFAULT 'Ayanda'` - AWS Polly voice ID
- `voice_engine VARCHAR(50) DEFAULT 'standard'` - Voice engine (standard or neural)

**Impact**: Allows per-project voice customization instead of using global config

## Creating New Migrations

1. Create two files:
   - `00X_description.up.sql` - Migration
   - `00X_description.down.sql` - Rollback

2. Follow naming convention:
   - Use sequential numbering (001, 002, 003, ...)
   - Use snake_case for description
   - Keep descriptions concise

3. Test both up and down migrations:
   ```bash
   psql $DATABASE_URL -f migrations/00X_description.up.sql
   psql $DATABASE_URL -f migrations/00X_description.down.sql
   ```

## Notes

- Migrations are run manually (no automatic migration tool currently)
- Always test migrations on a development database first
- Document any data transformations or breaking changes
- Include DEFAULT values for new columns to avoid affecting existing rows
