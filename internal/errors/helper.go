package errors

func GetCode(err error) int {
	type hasCode interface {
		Code() int
	}

	if err != nil {
		code, ok := err.(hasCode)
		if !ok {
			return 0
		}
		return code.Code()
	}
	return 0
}
