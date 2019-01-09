package goose

import (
	"database/sql"
	"fmt"
)

// SQLDialect abstracts the details of specific SQL dialects
// for goose's few SQL specific statements
type SQLDialect interface {
	CreateVersionTableSQL() string // sql string to create the db version table
	InsertVersionSQL() string      // sql string to insert the initial version table row
	DeleteVersionSQL() string      // sql string to delete version
	MigrationSQL() string          // sql string to retrieve migrations
	DBVersionQuery(db *sql.DB) (*sql.Rows, error)
}

var dialects = map[string]SQLDialect{
	"postgres": &PostgresDialect{},
	"mysql":    &MySQLDialect{},
	"sqlite3":  &Sqlite3Dialect{},
	"mssql":    &SqlServerDialect{},
	"redshift": &RedshiftDialect{},
	"tidb":     &TiDBDialect{},
}

var dialect = dialects["postgres"]

// GetDialect gets the SQLDialect
func GetDialect() SQLDialect {
	return dialect
}

// SetDialect sets the SQLDialect
func SetDialect(name string) error {
	d, ok := dialects[name]
	if !ok {
		return fmt.Errorf("%q: unknown dialect", name)
	}

	dialect = d
	return nil
}

// RegisterDialect registers a new SQLDialect implementation by name
func RegisterDialect(name string, d SQLDialect) {
	dialects[name] = d
}

// GetDialects returns the list of registered dialects
func GetDialects() []string {
	var names []string
	for name := range dialects {
		names = append(names, name)
	}
	return names
}

////////////////////////////
// Postgres
////////////////////////////

// PostgresDialect struct.
type PostgresDialect struct{}

func (pg PostgresDialect) CreateVersionTableSQL() string {
	return fmt.Sprintf(`CREATE TABLE %s (
            	id serial NOT NULL,
                version_id bigint NOT NULL,
                is_applied boolean NOT NULL,
                tstamp timestamp NULL default now(),
                PRIMARY KEY(id)
            );`, TableName())
}

func (pg PostgresDialect) InsertVersionSQL() string {
	return fmt.Sprintf("INSERT INTO %s (version_id, is_applied) VALUES ($1, $2);", TableName())
}

func (pg PostgresDialect) DBVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT version_id, is_applied from %s ORDER BY id DESC", TableName()))
	if err != nil {
		return nil, err
	}

	return rows, err
}

func (m PostgresDialect) MigrationSQL() string {
	return fmt.Sprintf("SELECT tstamp, is_applied FROM %s WHERE version_id=$1 ORDER BY tstamp DESC LIMIT 1", TableName())
}

func (pg PostgresDialect) DeleteVersionSQL() string {
	return fmt.Sprintf("DELETE FROM %s WHERE version_id=$1;", TableName())
}

////////////////////////////
// MySQL
////////////////////////////

// MySQLDialect struct.
type MySQLDialect struct{}

func (m MySQLDialect) CreateVersionTableSQL() string {
	return fmt.Sprintf(`CREATE TABLE %s (
                id serial NOT NULL,
                version_id bigint NOT NULL,
                is_applied boolean NOT NULL,
                tstamp timestamp NULL default now(),
                PRIMARY KEY(id)
            );`, TableName())
}

func (m MySQLDialect) InsertVersionSQL() string {
	return fmt.Sprintf("INSERT INTO %s (version_id, is_applied) VALUES (?, ?);", TableName())
}

func (m MySQLDialect) DBVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT version_id, is_applied from %s ORDER BY id DESC", TableName()))
	if err != nil {
		return nil, err
	}

	return rows, err
}

func (m MySQLDialect) MigrationSQL() string {
	return fmt.Sprintf("SELECT tstamp, is_applied FROM %s WHERE version_id=? ORDER BY tstamp DESC LIMIT 1", TableName())
}

func (m MySQLDialect) DeleteVersionSQL() string {
	return fmt.Sprintf("DELETE FROM %s WHERE version_id=?;", TableName())
}

////////////////////////////
// MSSQL
////////////////////////////

// SqlServerDialect struct.
type SqlServerDialect struct{}

func (m SqlServerDialect) CreateVersionTableSQL() string {
	return fmt.Sprintf(`CREATE TABLE %s (
                id INT NOT NULL IDENTITY(1,1) PRIMARY KEY,
                version_id BIGINT NOT NULL,
                is_applied BIT NOT NULL,
                tstamp DATETIME NULL DEFAULT CURRENT_TIMESTAMP
            );`, TableName())
}

func (m SqlServerDialect) InsertVersionSQL() string {
	return fmt.Sprintf("INSERT INTO %s (version_id, is_applied) VALUES (@p1, @p2);", TableName())
}

func (m SqlServerDialect) DBVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT version_id, is_applied FROM %s ORDER BY id DESC", TableName()))
	if err != nil {
		return nil, err
	}

	return rows, err
}

func (m SqlServerDialect) MigrationSQL() string {
	const tpl = `
WITH Migrations AS
(
    SELECT tstamp, is_applied,
    ROW_NUMBER() OVER (ORDER BY tstamp) AS 'RowNumber'
    FROM %s
	WHERE version_id=@p1
)
SELECT tstamp, is_applied
FROM Migrations
WHERE RowNumber BETWEEN 1 AND 2
ORDER BY tstamp DESC
`
	return fmt.Sprintf(tpl, TableName())
}

func (m SqlServerDialect) DeleteVersionSQL() string {
	return fmt.Sprintf("DELETE FROM %s WHERE version_id=@p1;", TableName())
}

////////////////////////////
// sqlite3
////////////////////////////

// Sqlite3Dialect struct.
type Sqlite3Dialect struct{}

func (m Sqlite3Dialect) CreateVersionTableSQL() string {
	return fmt.Sprintf(`CREATE TABLE %s (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                version_id INTEGER NOT NULL,
                is_applied INTEGER NOT NULL,
                tstamp TIMESTAMP DEFAULT (datetime('now'))
            );`, TableName())
}

func (m Sqlite3Dialect) InsertVersionSQL() string {
	return fmt.Sprintf("INSERT INTO %s (version_id, is_applied) VALUES (?, ?);", TableName())
}

func (m Sqlite3Dialect) DBVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT version_id, is_applied from %s ORDER BY id DESC", TableName()))
	if err != nil {
		return nil, err
	}

	return rows, err
}

func (m Sqlite3Dialect) MigrationSQL() string {
	return fmt.Sprintf("SELECT tstamp, is_applied FROM %s WHERE version_id=? ORDER BY tstamp DESC LIMIT 1", TableName())
}

func (m Sqlite3Dialect) DeleteVersionSQL() string {
	return fmt.Sprintf("DELETE FROM %s WHERE version_id=?;", TableName())
}

////////////////////////////
// Redshift
////////////////////////////

// RedshiftDialect struct.
type RedshiftDialect struct{}

func (rs RedshiftDialect) CreateVersionTableSQL() string {
	return fmt.Sprintf(`CREATE TABLE %s (
            	id integer NOT NULL identity(1, 1),
                version_id bigint NOT NULL,
                is_applied boolean NOT NULL,
                tstamp timestamp NULL default sysdate,
                PRIMARY KEY(id)
            );`, TableName())
}

func (rs RedshiftDialect) InsertVersionSQL() string {
	return fmt.Sprintf("INSERT INTO %s (version_id, is_applied) VALUES ($1, $2);", TableName())
}

func (rs RedshiftDialect) DBVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT version_id, is_applied from %s ORDER BY id DESC", TableName()))
	if err != nil {
		return nil, err
	}

	return rows, err
}

func (m RedshiftDialect) MigrationSQL() string {
	return fmt.Sprintf("SELECT tstamp, is_applied FROM %s WHERE version_id=$1 ORDER BY tstamp DESC LIMIT 1", TableName())
}

func (rs RedshiftDialect) DeleteVersionSQL() string {
	return fmt.Sprintf("DELETE FROM %s WHERE version_id=$1;", TableName())
}

////////////////////////////
// TiDB
////////////////////////////

// TiDBDialect struct.
type TiDBDialect struct{}

func (m TiDBDialect) CreateVersionTableSQL() string {
	return fmt.Sprintf(`CREATE TABLE %s (
                id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT UNIQUE,
                version_id bigint NOT NULL,
                is_applied boolean NOT NULL,
                tstamp timestamp NULL default now(),
                PRIMARY KEY(id)
            );`, TableName())
}

func (m TiDBDialect) InsertVersionSQL() string {
	return fmt.Sprintf("INSERT INTO %s (version_id, is_applied) VALUES (?, ?);", TableName())
}

func (m TiDBDialect) DBVersionQuery(db *sql.DB) (*sql.Rows, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT version_id, is_applied from %s ORDER BY id DESC", TableName()))
	if err != nil {
		return nil, err
	}

	return rows, err
}

func (m TiDBDialect) MigrationSQL() string {
	return fmt.Sprintf("SELECT tstamp, is_applied FROM %s WHERE version_id=? ORDER BY tstamp DESC LIMIT 1", TableName())
}

func (m TiDBDialect) DeleteVersionSQL() string {
	return fmt.Sprintf("DELETE FROM %s WHERE version_id=?;", TableName())
}
