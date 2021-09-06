package main

func noError(err error) {
	if err != nil {
		panic(err)
	}
}
