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
    msg_type            text    NOT NULL, -- Maps to MsgType
    trade_report_id     text    NOT NULL, -- Maps to TradeReportID
    exec_id             text    NOT NULL, -- Maps to ExecID
    order_id            text    NOT NULL, -- Maps to OrderID
    cl_ord_id           text,             -- Maps to ClOrdID
    symbol              text    NOT NULL, -- Maps to Symbol
    side                text    NOT NULL, -- Maps to Side
    last_qty            numeric NOT NULL, -- Maps to LastQty
    last_px             numeric NOT NULL, -- Maps to LastPx
    trade_date          text    NOT NULL, -- Maps to TradeDate
    transact_time       bigint  NOT NULL, -- Maps to TransactTime
    previously_reported boolean NOT NULL, -- Maps to PreviouslyReported
    text                text,             -- Maps to Text
    PRIMARY KEY (trade_report_id)
);