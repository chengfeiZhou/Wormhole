package logger

var (
	ErrorParam            = AppError{code: 1001, msg: "Invalid Params"}
	ErrorNonExistsFolder  = AppError{code: 1002, msg: "Folder does not exists"}
	ErrorNonExistsFile    = AppError{code: 1003, msg: "File does not exists"}
	ErrorEmptyData        = AppError{code: 1004, msg: "Data is empty"}
	ErrorMethodNotSupport = AppError{code: 1005, msg: "Method does not support"}
	ErrorMethod           = AppError{code: 1006, msg: "Method execution exception"}
	ErrorInstantiation    = AppError{code: 1007, msg: "Instance creation exception"}
	ErrorWriteFile        = AppError{code: 1008, msg: "Write file exception"}
	ErrorReadFile         = AppError{code: 1009, msg: "Read file exception"}

	ErrorRouter          = AppError{code: 3001, msg: "Router execution exception"} // API Router
	ErrorRouterRegister  = AppError{code: 3002, msg: "Router Register exception"}
	ErrorAgentStart      = AppError{code: 3003, msg: "Agent Start exception"}
	ErrorMakeMiddleware  = AppError{code: 3004, msg: "Make Middleware exception"}
	ErrorHandleError     = AppError{code: 3005, msg: "Error Handle exception"}
	ErrorRequestExecutor = AppError{code: 3006, msg: "HTTP Request Executor exception"}
	ErrorServerPlugin    = AppError{code: 3007, msg: "ServerPlugin exception"}

	ErrorParamsIncomplete     = AppError{code: 4001, msg: "Incomplete parameters"} // 参数
	ErrorAuthParamsIncomplete = AppError{code: 4101, msg: "Register/UnRegister body is null"}

	ErrorHost          = AppError{code: 5001, msg: "Request address exception"} // 网络请求错误
	ErrorHTTPHandle    = AppError{code: 5002, msg: "HTTP Handle exception"}
	ErrorNetForwarding = AppError{code: 5003, msg: "Network forwarding error"}
	ErrorMiddleware    = AppError{code: 5004, msg: "Middleware Handle error"}

	ErrorKafka             = AppError{code: 6001, msg: "Kafka execution exception"} // kafka
	ErrorKafkaConsumer     = AppError{code: 6101, msg: "Kafka consumer execution exception"}
	ErrorKafkaProducer     = AppError{code: 6201, msg: "Kafka producer execution exception"}
	ErrorKafkaProducerSend = AppError{code: 6202, msg: "Kafka producer send data execution exception"}
	// 可扩展...
)

// AppError 业务逻辑错误对象
type AppError struct {
	msg  string
	code int
}

// Code 返回错误代码
func (r AppError) Code() int {
	return r.code
}

// Error 返回错误信息
func (r AppError) Error() string {
	return r.msg
}

// IsAppError 检查给定的错误是否为AppError类型
func IsAppError(err error) bool {
	_, ok := err.(AppError)
	return ok
}
