CREATE USER cryptellation;
ALTER USER cryptellation PASSWORD 'cryptellation';
ALTER USER cryptellation CREATEDB;

CREATE DATABASE candlesticks;
GRANT ALL PRIVILEGES ON DATABASE candlesticks TO cryptellation;
\c candlesticks postgres
GRANT ALL ON SCHEMA public TO cryptellation;

CREATE DATABASE sma;
GRANT ALL PRIVILEGES ON DATABASE sma TO cryptellation;
\c sma postgres
GRANT ALL ON SCHEMA public TO cryptellation; 