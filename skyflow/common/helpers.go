package common

func AppendRequestId(message string, requestId string) string {
	if requestId == "" {
		return message
	}

	return message + " - requestId : " + requestId
}
