FROM postgres:15-alpine

# Copy initialization script
COPY init.sql /docker-entrypoint-initdb.d/

# The postgres image automatically runs scripts in /docker-entrypoint-initdb.d/
# on first container startup when the database is initialized
