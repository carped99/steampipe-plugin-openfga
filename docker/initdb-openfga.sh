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

# Optionally create default foreign server if environment variables are set
#if [ -n "$OPENFGA_ENDPOINT" ] && [ -n "$OPENFGA_STORE_ID" ]; then
#	echo "Creating default OpenFGA foreign server"
#
#	# Build OPTIONS clause
#	OPTIONS="endpoint '${OPENFGA_ENDPOINT}', store_id '${OPENFGA_STORE_ID}'"
#
#	# Add optional parameters
#	if [ -n "$OPENFGA_API_TOKEN" ]; then
#		OPTIONS="${OPTIONS}, api_token '${OPENFGA_API_TOKEN}'"
#	fi
#
#	if [ -n "$OPENFGA_AUTHORIZATION_MODEL_ID" ]; then
#		OPTIONS="${OPTIONS}, authorization_model_id '${OPENFGA_AUTHORIZATION_MODEL_ID}'"
#	fi
#
#	psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
#		-- Create default foreign server
#		CREATE SERVER IF NOT EXISTS openfga_server
#		  FOREIGN DATA WRAPPER openfga_fdw
#		  OPTIONS (${OPTIONS});
#
#		-- Create user mapping for postgres user
#		CREATE USER MAPPING IF NOT EXISTS FOR ${POSTGRES_USER}
#		  SERVER openfga_server;
#
#		-- Import foreign schema (optional)
#		-- IMPORT FOREIGN SCHEMA openfga
#		--   FROM SERVER openfga_server
#		--   INTO public;
#	EOSQL
#
#	echo "Default OpenFGA server created successfully"
#else
#	echo "Skipping foreign server creation (set OPENFGA_ENDPOINT and OPENFGA_STORE_ID to auto-create)"
#fi
