package mockserver

type Request struct {
	Method string `json:"method"`
	Path string `json:"path"`
	QueryStringParameters []*NameValues `json:"queryStringParameters,omitempty"`
	Headers []*NameValues `json:"headers,omitempty"`
	Cookies []*NameValues `json:"cookies,omitempty"`
	Body *Body `json:"body,omitempty"`
}

func NewRequest(method, path string) *Request {
	return &Request{
		Method: method,
		Path: path,
		QueryStringParameters: make([]*NameValues, 0),
		Headers: make([]*NameValues, 0),
		Cookies: make([]*NameValues, 0),
	}
}

func (r *Request) AddQueryStringParameter(name, value string) *Request {
	addNameValue(&r.QueryStringParameters, name, value)
	return r
}

func (r *Request) AddHeader(name, value string) *Request {
	addNameValue(&r.Headers, name, value)
	return r
}

func (r *Request) AddCookie(name, value string) *Request {
	addNameValue(&r.Cookies, name, value)
	return r
}

func (r *Request) SetStringBody(body string) *Request {
	r.Body = &Body{
		Type: "STRING",
		Value: body,
	}
	return r
}

func (r *Request) SetJSONBody(jsonBody string) *Request {
	r.Body = &Body{
		Type: "JSON",
		Value: jsonBody,
	}
	return r
}

func (r *Request) SetJSONBodyWithMatchType(matchType, jsonBody string) *Request {
	r.Body = &Body{
		Type: "JSON",
		Value: jsonBody,
		MatchType: matchType,
	}
	return r
}

func addNameValue(nameValues *[]*NameValues, name, value string) {
	for _, p := range *nameValues {
		if p.Name == name {
			p.AddNameValue(name, value)
			return
		}
	}

	newNameValues := append(
		*nameValues,
		&NameValues{
			Name: name,
			Value: value,
		})
	nameValues = &newNameValues
}

type Response struct {
	StatusCode int `json:"statusCode"`
	Headers []*NameValues `json:"headers,omitempty"`
	Cookies []*NameValues `json:"cookies,omitempty"`
	Body string `json:"body,omitempty"`
	Delay *Delay `json:"delay,omitempty"`
}

func NewResponse(statusCode int) *Response {
	return &Response{
		StatusCode: statusCode,
		Headers: make([]*NameValues, 0),
		Cookies: make([]*NameValues, 0),
	}
}

func (r *Response) AddHeader(name, value string) *Response {
	addNameValue(&r.Headers, name, value)
	return r
}

func (r *Response) AddCookie(name, value string) *Response {
	addNameValue(&r.Cookies, name, value)
	return r
}

func (r *Response) SetBody(body string) *Response {
	r.Body = body
	return r
}

func (r *Response) SetDelay(timeUnit string, value float64) *Response {
	r.Delay = &Delay{
		TimeUnit: timeUnit,
		Value: value,
	}
	return r
}

type NameValues struct {
	Name string `json:"name"`
	Value string `json:"value,omitempty"`
	Values []string `json:"values,omitempty"`
}

func (kv *NameValues) AddNameValue(name, value string) *NameValues {
	if len(kv.Value) > 0 {
		kv.Values = append(kv.Values, kv.Value, value)
		kv.Value = ""
	} else if len(kv.Values) > 0 {
		kv.Values = append(kv.Values, value)
	} else {
		kv.Value = value
	}
	return kv
}

type Delay struct {
	TimeUnit string `json:"timeUnit"` // SECONDS, MINUTES...
	Value float64 `json:"value"`
}

type Body struct {
	Type string `json:"type"` // STRING or JSON
	Value string `json:"value"`
	MatchType string `json:"matchType,omitempty"`
}

type MockTimes struct {
	RemainingTimes int `json:remainingTimes"`
	Unlimited bool `json:"unlimited"`
}

type TimeToLive struct {
	TimeUnit string `json:"timeUnit"` // SECONDS, MINUTES...
	TimeToLive float64 `json:"timeToLive"`
	Unlimited bool `json:"unlimited"`
}

type MockAnyResponse struct {
	HttpRequest *Request `json:"httpRequest"`
	HttpResponse *Response `json:"httpResponse"`
	Times *MockTimes `json:"times,omitempty"`
	TimeToLive *TimeToLive `json:"timeToLive,omitempty"`
}

func NewMockAnyResponse() *MockAnyResponse {
	return &MockAnyResponse{}
}

func (m *MockAnyResponse) When(request *Request) *MockAnyResponse {
	m.HttpRequest = request
	return m
}

func (m *MockAnyResponse) Respond(response *Response) *MockAnyResponse {
	m.HttpResponse = response
	return m
}

func (m *MockAnyResponse) WithTimes(remainingTimes int) *MockAnyResponse {
	m.Times = &MockTimes{
		RemainingTimes: remainingTimes,
		Unlimited: false,
	}
	return m
}

func (m *MockAnyResponse) WithTimeToLive(timeUnit string, timeToLive float64) *MockAnyResponse {
	m.TimeToLive = &TimeToLive{
		TimeUnit: timeUnit,
		TimeToLive: timeToLive,
		Unlimited: false,
	}
	return m
}

type Verify struct {
	HttpRequest *Request `json:"httpRequest"`
	Times *ProxyTimes `json:"times"`
}

func NewVerify() *Verify {
	return &Verify{}
}

func (v *Verify) MatchRequest(request *Request) *Verify {
	v.HttpRequest = request
	return v
}

func (v *Verify) WithTimes(count int, exact bool) *Verify {
	v.Times = &ProxyTimes{
		Count: count,
		Exact: exact,
	}
	return v
}

type ProxyTimes struct {
	Count int `json:"count"`
	Exact bool `json:"exact"`
}

type Retrieve struct {
	HttpRequest *Request `json:"httpRequest"`
}

func NewRetrieve() *Retrieve {
	return &Retrieve{}
}

func (r *Retrieve) MatchRequest(request *Request) *Retrieve {
	r.HttpRequest = request
	return r
}

type RetrievedRequest struct {
	Method string `json:"method"`
	Path string `json:"path"`
	Headers []*NameValues `json:"headers"`
	KeepAlive bool `json:"keepAlive"`
	Secure bool `json:"secure"`
	Body []byte `json:"body,omitempty"`
}