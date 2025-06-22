package model

type ExecType string

const (
	ExecTypeNew      ExecType = "0"
	ExecTypeFill     ExecType = "2"
	ExecTypeCanceled ExecType = "4"
	ExecTypeRejected ExecType = "8"
)

type OrderStatus string

const (
	OrderStatusNew         OrderStatus = "0"
	OrderStatusFill        OrderStatus = "2"
	OrderStatusPartialFill OrderStatus = "1"
	OrderStatusCanceled    OrderStatus = "4"
	OrderStatusRejected    OrderStatus = "8"
	OrderStatusPendingNew  OrderStatus = "A"
)
