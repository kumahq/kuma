package error

func MustNot(err error) {
	if err != nil {
		panic(err)
	}
}
