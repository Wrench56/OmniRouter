package logger

var callerModule string

func SetLogCallerModule(module string) {
	callerModule = module
}

func ConsumeLogCallerModule() string {
	m := callerModule
	callerModule = ""
	return m
}

func IsLogCallerModuleSet() bool {
	return callerModule != ""
}
