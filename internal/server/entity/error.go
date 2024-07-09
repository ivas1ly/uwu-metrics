package entity

import "errors"

var (
	ErrIncorrectMetricValue = errors.New("incorrect metric value")
	ErrUnknownMetricType    = errors.New("unknown metric type")
	ErrEmptyMetricValue     = errors.New("empty metric value")
	ErrCanNotGetMetricValue = errors.New("can't get metric value")
)
