package b

func E() error { // want E:`isTypedFunc\[0\]`
	var err *struct{error}
	return err
}

func NE() error {
	return nil
}
