CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE executions
(
    exec_id       text    NOT NULL, -- Maps to ExecID
    order_id      text    NOT NULL, -- Maps to OrderID
    cl_ord_id     text,             -- Maps to ClOrdID
    exec_type     text    NOT NULL, -- Maps to ExecType
    ord_status    text    NOT NULL, -- Maps to OrdStatus
    symbol        text    NOT NULL, -- Maps to Symbol
    side          text    NOT NULL, -- Maps to Side
    order_qty     numeric NOT NULL, -- Maps to OrderQty
    last_shares   numeric NOT NULL, -- Maps to LastShares
    last_px       numeric NOT NULL, -- Maps to LastPx
    leaves_qty    numeric NOT NULL, -- Maps to LeavesQty
    cum_qty       numeric NOT NULL, -- Maps to CumQty
    avg_px        numeric NOT NULL, -- Maps to AvgPx
    transact_time bigint  NOT NULL, -- Maps to TransactTime
    text          text,             -- Maps to Text
    PRIMARY KEY (exec_id)
);

CREATE TABLE trade_capture_reports
(
    trade_report_id text PRIMARY KEY, -- 571
    msg_type        text    NOT NULL, -- 35 (AE)
    exec_id         text    NOT NULL, -- 17
    symbol          text    NOT NULL, -- 55
    last_qty        NUMERIC NOT NULL, -- 32
    last_px         NUMERIC NOT NULL, -- 31
    trade_date      text    NOT NULL, -- 75 (YYYYMMDD)
    transact_time   bigint  NOT NULL  -- 60
);

CREATE TABLE trade_sides
(
    id              SERIAL PRIMARY KEY,
    trade_report_id text     NOT NULL REFERENCES trade_capture_reports (trade_report_id) ON DELETE CASCADE,
    side            SMALLINT NOT NULL, -- 54: 1 = Buy, 2 = Sell
    order_id        text     NOT NULL  -- 37
);
