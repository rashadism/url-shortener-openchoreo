FROM postgres:15-alpine

# Copy initialization script
COPY init.sql /docker-entrypoint-initdb.d/

# Fix permissions so postgres user can read the init script
RUN chmod 644 /docker-entrypoint-initdb.d/init.sql

# The postgres image automatically runs scripts in /docker-entrypoint-initdb.d/
# on first container startup when the database is initialized
