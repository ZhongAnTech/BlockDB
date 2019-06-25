package message

import "fmt"

type OpCode uint32

const (
	OpReply        = OpCode(1)
	OpUpdate       = OpCode(2001)
	OpInsert       = OpCode(2002)
	Reserved       = OpCode(2003)
	OpQuery        = OpCode(2004)
	OpGetMore      = OpCode(2005)
	OpDelete       = OpCode(2006)
	OpKillCursors  = OpCode(2007)
	OpCommand      = OpCode(2010)
	OpCommandReply = OpCode(2011)
	OpMsg          = OpCode(2013)
)

func (op *OpCode) String() string {
	switch *op {
	case OpReply:
		return "OpReply"
	case OpUpdate:
		return "OpUpdate"
	case OpInsert:
		return "OpInsert"
	case Reserved:
		return "Reserved"
	case OpQuery:
		return "OpQuery"
	case OpGetMore:
		return "OpGetMore"
	case OpDelete:
		return "OpDelete"
	case OpKillCursors:
		return "OpKillCursors"
	case OpCommand:
		return "OpCommand"
	case OpCommandReply:
		return "OpCommandReply"
	case OpMsg:
		return "OpMsg"
	default:
		return fmt.Sprintf("unknown op %d", *op)
	}
}
