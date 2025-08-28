package entity

import "github.com/Amierza/chat-service/constants"

type (
	Sender     string
	Identifier string
)

const (
	// message sender type
	USER Sender = constants.ENUM_SENDER_USER
	BOT  Sender = constants.ENUM_SENDER_BOT
)

func IsValidSender(s Sender) bool {
	return s == USER || s == BOT
}
