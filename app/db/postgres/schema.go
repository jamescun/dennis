package postgres

const (
	// migrationTable is the `CREATE TABLE` statement to create the
	// `migrations` table within PostgreSQL.
	//
	// NOTE(jc): schema migrations are not currently implemented, but this
	// table is created for future proofing when they do get implemented.
	migrationTable = `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version  INTEGER  NOT NULL,
			dirty    BOOLEAN  NOT NULL
		);
	`

	// queryTable is the `CREATE TABLE` statement to create the `queries` table
	// within PostgreSQL.
	queryTable = `
		CREATE TABLE IF NOT EXISTS queries (
			id    UUID  PRIMARY KEY DEFAULT uuidv7(),
			type  TEXT  NOT NULL,
			name  TEXT  NOT NULL,

			created_at   TIMESTAMPTZ  NOT NULL DEFAULT (now() at time zone 'UTC'),
			finished_at  TIMESTAMPTZ
		);
	`

	// lookupTable is the `CREATE TABLE` statement to create the `lookups`
	// table within PostgreSQL.
	lookupTable = `
		CREATE TABLE IF NOT EXISTS lookups (
			id        UUID  PRIMARY KEY DEFAULT uuidv7(),
			query_id  UUID  NOT NULL REFERENCES queries(id),

			resolver  TEXT     NOT NULL,
			rtt       INTEGER  NOT NULL,
			error     TEXT,

			resolved_at  TIMESTAMPTZ
		);

		CREATE INDEX IF NOT EXISTS lookups_query_id_idx
			ON lookups(query_id);
	`

	// recordTable is the `CREATE TABLE statement to create the `records`
	// table within PostgreSQL.
	recordTable = `
		CREATE TABLE IF NOT EXISTS records (
			lookup_id  UUID  NOT NULL REFERENCES lookups(id),

			ttl       INTEGER  NOT NULL,
			priority  INTEGER,
			weight    INTEGER,
			port      INTEGER,
			tag       TEXT,
			content   TEXT[]   NOT NULL
		);

		CREATE INDEX IF NOT EXISTS records_lookup_id_idx
			ON records(lookup_id);
	`
)
