CREATE USER cryptellation_sma;
ALTER USER cryptellation_sma PASSWORD 'cryptellation_sma';
ALTER USER cryptellation_sma CREATEDB;

CREATE DATABASE cryptellation_sma;
GRANT ALL PRIVILEGES ON DATABASE cryptellation_sma TO cryptellation_sma;
\c cryptellation_sma postgres
GRANT ALL ON SCHEMA public TO cryptellation_sma;