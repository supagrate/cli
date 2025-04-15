![npm](https://img.shields.io/npm/v/supagrate)


# Supagrate

Hasura-style [Supabase](https://supabase.com/) migrations.

## Installation

You can download the binaries for each versions, or install it in your project via npm:

```bash
npm install --save supagrate
```

### Requirements

- Docker (for schema diffing)

## Commands

| Command                   | Description                                        |
|---------------------------|----------------------------------------------------|
| `supagrate init`          | Setup your project with supagrate                  |
| `supagrate migrate diff`  | Show schema differences and create migrations      |
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
postgresql://postgres:postgres@localhost:5432/postgres?sslmode=disable
```

You can override this by setting the following environment variable: 

- `DATABASE_URL`: The full connection string to Postgres

You can also use a `.env` with the DATABASE_URL in it. 

Or by setting the following flags:

```bash
supagrate migrate
```

## CI

You can use Supagrate in GitHub Actions to run your migrations and seeders. 

Check out our [demo](https://github.com/supagrate/demo) repository for more info.