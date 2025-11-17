#!/bin/bash
set -e

# Create the 'template_openfga' template db
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE template_openfga IS_TEMPLATE true;
EOSQL

# Load OpenFGA FDW extension into both template_database and $POSTGRES_DB
for DB in template_openfga "$POSTGRES_DB"; do
	echo "Loading OpenFGA FDW extension into $DB"
	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$DB" <<-EOSQL
		CREATE EXTENSION IF NOT EXISTS openfga_fdw;
	EOSQL
done
