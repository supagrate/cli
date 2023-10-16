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

## Roadmap

- [ ] Use already setup [Supabase CLI](https://github.com/supabase/cli) to login to the DB
- [ ] Allow 0 setup for remote DB (to be able to run migrations on GitHub Actions)
- [ ] Assume local DB is `postgresql://postgres:postgres@localhost:54322/postgres?sslmode=disable`
- [x] Allow the generation of migrations
- [x] Use `up.sql` and `down.sql` files for migrations
- [x] Allow complete reset of the DB
- [ ] Allow the application (up) of migrations
- [ ] Allow the rollback of migrations
- [ ] Allow the generation of arbitrary seeder
- [ ] Allow the application of arbitrary seeder
