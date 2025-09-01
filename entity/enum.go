package entity

import "github.com/Amierza/chat-service/constants"

type (
	Role          string
	Degree        string
	Progress      string
	SessionStatus string
)

const (
	STUDENT  Role = constants.ENUM_ROLE_STUDENT
	LECTURER Role = constants.ENUM_ROLE_LECTURER

	S1 Degree = constants.ENUM_DEGREE_S1
	S2 Degree = constants.ENUM_DEGREE_S2
	S3 Degree = constants.ENUM_DEGREE_S3

	BAB1 Progress = constants.ENUM_PROGRESS_BAB1
	BAB2 Progress = constants.ENUM_PROGRESS_BAB2
	BAB3 Progress = constants.ENUM_PROGRESS_BAB3

	WAITING  SessionStatus = constants.ENUM_SESSION_STATUS_WAITING
	ONGOING  SessionStatus = constants.ENUM_SESSION_STATUS_ONGOING
	FINISHED SessionStatus = constants.ENUM_SESSION_STATUS_FINSIHED
)

func IsValidRole(r Role) bool {
	return r == STUDENT || r == LECTURER
}
func IsValidDegree(d Degree) bool {
	return d == S1 || d == S2 || d == S3
}
func IsValidProgress(p Progress) bool {
	return p == BAB1 || p == BAB2 || p == BAB3
}
func IsValidSessionStatus(ss SessionStatus) bool {
	return ss == WAITING || ss == ONGOING || ss == FINISHED
}
