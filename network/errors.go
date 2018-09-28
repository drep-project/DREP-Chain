package network

type myError struct {
    Err error
}

func (e *myError) Error() string {
    return e.Err.Error()
}

type DataError struct {
    myError
}

type TimeoutError struct {
    myError
}

type ConnectionError struct {
    myError
}

type TransmissionError struct {
    myError
}

type DefaultError struct {
    myError
}