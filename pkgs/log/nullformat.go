package log

import "github.com/sirupsen/logrus"

//NullFormat to disable default output
type NullFormat struct {
}

func (format *NullFormat) Format(*logrus.Entry) ([]byte, error) {
	return nil, nil
}
