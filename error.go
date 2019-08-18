package mint

import "runtime"

//Error is context for error
type Error struct {
	error
	file     string
	line     int
	funcName string
}

func (err Error) Error() string {
	return err.error.Error()
}

//Traceable error stores error with context of error like function name
func Traceable(err error) Error {
	var errObj Error
	if err == nil {
		return errObj
	}
	switch err.(type) {
	case Error:
		return err.(Error)
	}
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)
	f := runtime.FuncForPC(pc[0])
	file, line := f.FileLine(pc[0])
	errObj = Error{
		error:    err,
		file:     file,
		line:     line,
		funcName: f.Name(),
	}
	return errObj
}
