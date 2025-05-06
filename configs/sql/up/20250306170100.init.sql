CREATE TABLE sma
(
    exchange VARCHAR(100) NOT NULL,
    pair VARCHAR(100) NOT NULL,
    period VARCHAR(100) NOT NULL,
    period_number INTEGER NOT NULL,
    price_type VARCHAR(100) NOT NULL,
    time TIMESTAMP NOT NULL,
    data JSONB NOT NULL,
    CONSTRAINT pk_sma PRIMARY KEY (exchange, pair, period, period_number, price_type, time)
);