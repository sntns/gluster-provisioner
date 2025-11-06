package device

// LogicalOp defines the logical operation for chaining filters.
type LogicalOp string

const (
	OpAnd LogicalOp = "AND"
	OpOr  LogicalOp = "OR"
)
