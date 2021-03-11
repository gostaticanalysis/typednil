package a

import "a/b"

func f1() {
	err := e()
	if err != nil { // want "it may become a comparition a typed nil and an untyped nil"
		print(err)
	}
}

func f2() {
	err := b.E()
	if err != nil { // want "it may become a comparition a typed nil and an untyped nil"
		print(err)
	}
}

func f3() {
	err := b.NE()
	if err != nil { // OK
		print(err)
	}
}

func e() error { // want e:`isTypedFunc\[0\]`
	var err *struct{error}
	return err
}
