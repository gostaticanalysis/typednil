package b

type MyError struct{ error }

func E() error { // want E:`nilable results \[0:I\]`
	var err *MyError
	return err
}

func NE() error {
	return nil
}

func CE1() (int, *MyError) { // want CE1:`nilable results \[1:C\]`
	return 0, nil
}

// CE2 does not return nil.
func CE2() (int, *MyError) {
	return 0, new(MyError)
}
