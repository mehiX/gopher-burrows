package burrows

type requestType string

const (
	ReqStatus    requestType = "status"
	ReqAvailable requestType = "available"
	ReqGopher    requestType = "gopher"
	ReqClose     requestType = "close"
)

// Response from a burrow to the manager.
// Should contain the current status and may also contain a new request channel
// if further instructions are expected from the manager
type Response struct {
	burrow      Burrow
	nextRequest chan Request
}

// Request is a request from the manager to a burrow.
// It provides a channel where the burrow can send its response
type Request struct {
	name     requestType
	response chan Response
}

func NewStatusRequest(resp chan Response) Request {
	return Request{
		name:     ReqStatus,
		response: resp,
	}
}

func NewAvailableRequest() Request {
	return Request{
		name:     ReqAvailable,
		response: make(chan Response, 1), // we are only interested in the first available burrow, the rest of the responses are discarded.
	}
}

func NewGopherRequest() Request {
	return Request{
		name:     ReqGopher,
		response: make(chan Response),
	}
}
