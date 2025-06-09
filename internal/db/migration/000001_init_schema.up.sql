CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE active_orders
(
    id           text    NOT NULL,
    side         text    NOT NULL,
    order_qty    numeric NOT NULL,
    leaves_qty   numeric NOT NULL,
    price        numeric NOT NULL,
    instrument   text    NOT NULL,
    type         text    NOT NULL,
    exec_qty     numeric NOT NULL,
    order_status text    NOT NULL,
    PRIMARY KEY (id)
);