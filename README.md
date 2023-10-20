![npm](https://img.shields.io/npm/v/supagrate)


# Supagrate

The missing tool for [Supabase](https://supabase.com/) migrations.

## Installation

[//]: # (@TODO)

## Commands

| Command                   | Description                                        |
|---------------------------|----------------------------------------------------|
| `supagrate init`          | Setup your project with supagrate                  |
| `supagrate migrate down`  | Rollback your migrations one by one                |
| `supagrate migrate new`   | Generate a new migration                           |
| `supagrate migrate reset` | Reset your database and apply all migrations again |
| `supagrate migrate up`    | Apply your migrations one by one                   |
| `supagrate seed apply`    | Apply a seeder to your database                    |
| `supagrate seed new`      | Generate a new seeder                              |
| `supagrate status`        | View the migration status of your database         |

## Database Configuration

For the `supagrate migrate` and `supabase seed` commands, the default is to use the local Supabase database: 

```
postgresql://postgres:postgres@localhost:54322/postgres?sslmode=disable
```

You can override this by setting the following environment variables: 

- `DB_HOST`: The host of the database (ex. db.123456.supabase.co)
- `DB_PORT`: The port of the database (ex. 5432)
- `DB_USER`: The user of the database (ex. postgres)
- `DB_PASSWORD`: The password of the database (ex. postgres)
- `DB_NAME`: The name of the database (ex. postgres)

Or by setting the following flags:

```bash
supagrate migrate --db-host db.123456.supabase.co --db-port 5432 --db-user postgres --db-password postgres --db-name postgres
```

## CI

You can use Supagrate in GitHub Actions to run your migrations and seeders. 

Check out our [demo](https://github.com/supagrate/demo) repository for more info.