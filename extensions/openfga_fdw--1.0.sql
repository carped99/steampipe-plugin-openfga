/* openfga_fdw--1.0.sql */

-- complain if script is sourced in psql, rather than via CREATE EXTENSION
\echo Use "CREATE EXTENSION openfga_fdw" to load this extension. \quit

CREATE FUNCTION openfga_fdw_handler()
    RETURNS fdw_handler
    AS 'MODULE_PATHNAME', 'steampipe_openfga_fdw_handler'
    LANGUAGE C STRICT;

CREATE FUNCTION openfga_fdw_validator(text[], oid)
    RETURNS void
    AS 'MODULE_PATHNAME', 'steampipe_openfga_fdw_validator'
    LANGUAGE C STRICT;

CREATE FOREIGN DATA WRAPPER openfga_fdw
    HANDLER openfga_fdw_handler
    VALIDATOR openfga_fdw_validator;
