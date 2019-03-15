package health

type OutputFormat string

func (o OutputFormat) String() string {
	return string(o)
}

const (
	OutputProto OutputFormat = "proto"
	OutputJSON               = "json"
	OutputMesg               = "mesg"
)

const (
	ServiceKey      = "service"
	OutputFormatKey = "format"
)

const (
	StdRoute = "/health"
)
