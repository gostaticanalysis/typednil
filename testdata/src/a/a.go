package a

import "a/b"

func f1() {
	err := e()
	if err != nil { // want "typed nil"
		print(err)
	}
}

func f2() {
	err := b.E()
	if err != nil { // want "typed nil"
		print(err)
	}
}

func e() error { // want e:`isTypedFunc\[0\]`
	var err *struct{error}
	return err
}
