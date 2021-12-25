package common

func NoError(err error) {
	if err != nil {
		panic(err)
	}
}

func DrainChannel(ch <-chan error) {
	for range ch {
	}
}
