CREATE
EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE events
(
    id         uuid                 DEFAULT uuid_generate_v4(),
    order_id   text        NOT NULL,
    timestamp  timestamptz NOT NULL DEFAULT CURRENT_TIMESTAMP,
    type       text        NOT NULL,
    side       text        NOT NULL,
    instrument text        NOT NULL,
    order_qty  numeric,
    leaves_qty numeric,
    exec_qty   numeric,
    price      numeric,
    PRIMARY KEY (id)
);

CREATE TABLE active_orders
(
    id         text    NOT NULL,
    side       text    NOT NULL,
    qty        numeric NOT NULL,
    leaves_qty numeric NOT NULL,
    price      numeric NOT NULL,
    instrument  text    NOT NULL,
    PRIMARY KEY (id)
);