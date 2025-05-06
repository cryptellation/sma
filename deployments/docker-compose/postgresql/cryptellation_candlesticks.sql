CREATE USER cryptellation_candlesticks;
ALTER USER cryptellation_candlesticks PASSWORD 'cryptellation_candlesticks';
ALTER USER cryptellation_candlesticks CREATEDB;

CREATE DATABASE cryptellation_candlesticks;
GRANT ALL PRIVILEGES ON DATABASE cryptellation_candlesticks TO cryptellation_candlesticks;
\c cryptellation_candlesticks postgres
GRANT ALL ON SCHEMA public TO cryptellation_candlesticks;