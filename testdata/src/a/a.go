package a

import "a/b"

func e() error { // want e:`nilable results \[0:I\]`
	var err *struct{ error }
	return err
}

func f1() {
	err := e()
	if err != nil { // want "it may become a comparison between a typed nil and an untyped nil"
		print(err)
	}
	if err == nil { // want "it may become a comparison between a typed nil and an untyped nil"
		print(err)
	}
}

func f2() {
	err := b.E()
	if err != nil { // want "it may become a comparison between a typed nil and an untyped nil"
		print(err)
	}
	if err == nil { // want "it may become a comparison between a typed nil and an untyped nil"
		print(err)
	}
}

func f3() {
	err := b.NE()
	if err != nil { // OK
		print(err)
	}
}

func f4() {
	var err error
	_, err = b.CE1()
	if err != nil { // want "it may become a comparison between a typed nil and an untyped nil"
		print(err)
	}
	if err == nil { // want "it may become a comparison between a typed nil and an untyped nil"
		print(err)
	}
}

func f5() {
	_, err := b.CE1()
	if err != nil { // OK
		print(err)
	}
}

func f6() {
	var err error
	_, err = b.CE2()
	if err != nil { // OK
		print(err)
	}
}

func f7() {
	err := b.NE2()
	if err != nil { // OK
		print(err)
	}
}
