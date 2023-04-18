package constants

type RespStatus int32

const (
	ResOk RespStatus = iota
	ResErr
	ResNx
)
