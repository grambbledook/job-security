package asserts

import "fmt"

func NoError(err error, msg string) {
	if err != nil {
		panic(fmt.Sprintf("NoError assertion failed: [%s] - %v", msg, err))
	}
}

func NotEmpty(val, msg string) {
	if val == "" {
		panic(fmt.Sprintf("NotEmpty assertion failed: [%s]", msg))
	}
}

func NotZeru(val int, msg string) {
	if val == 0 {
		panic(fmt.Sprintf("NotEmpty assertion failed: [%s]", msg))
	}
}
