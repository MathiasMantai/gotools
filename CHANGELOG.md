## v0.2.1 (2025-09-02)
- created a wrapper for migrations for all supported database types
- created migrations for mysql and sqlite

## v0.2.0 (2025-08-19)

### db Package
- created an interface GotoolsDb which all available database types implement
- created a struct Db with Methods Query, QueryRow, BeginTx and Exec to use as a wrapper for all database types

## v0.1.2 (2025-07-01)

### Bugfixes
- fixed a bug in the env package that triggered an error when spaces are found inside an env file

### Other changes
- removed the language package and most the function GetSystemLanguage to the osutil package
- cleanup of multiple packages