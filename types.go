package health

type OutputFormat string

func (o OutputFormat) String() string {
	return string(o)
}

const (
	StdRoute        = "/health"
	ServiceKey      = "service"
	OutputFormatKey = "format"
)

const (
	OutputProto OutputFormat = "proto"
	OutputJSON               = "json"
	OutputMesg               = "mesg"
)
