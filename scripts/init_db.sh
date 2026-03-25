#!/bin/bash
set -e
 
echo "Creating service databases..."
 
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" <<-EOSQL
    CREATE DATABASE identity_db;
    CREATE DATABASE catalog_db;
    CREATE DATABASE commerce_db;
EOSQL
 
echo "Databases created successfully."
