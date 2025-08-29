package entity

import "github.com/Amierza/chat-service/constants"

type (
	Role   string
	Sender string
)

const (
	// message role type
	STUDENT  Role = constants.ENUM_ROLE_STUDENT
	LECTURER Role = constants.ENUM_ROLE_LECTURER

	// message sender type
	USER Sender = constants.ENUM_SENDER_USER
	BOT  Sender = constants.ENUM_SENDER_BOT
)

func IsValidSender(s Sender) bool {
	return s == USER || s == BOT
}
