package twocaptcha

import "github.com/valyala/fasthttp"

// requests.go contains function calls for compactly performing HTTP GET and POST requests
// Functional options included for attaching headers, skipping request body, etc.

// optHeaders attaches the passed headers (in format map[string]string) to the request.
func optHeaders(headers map[string]string) func(*fasthttp.Request, *fasthttp.Response) {
	return func(request *fasthttp.Request, response *fasthttp.Response) {
		for key, value := range headers {
			request.Header.Set(key, value)
		}
	}
}

// compactGET performs a HTTP GET request using the provided client.
// Returns the response (note: must be externally released) if request successful, else returns err != nil.
func compactGET(
	httpClient *fasthttp.Client, uri string, options ...func(*fasthttp.Request, *fasthttp.Response),
) (response *fasthttp.Response, finalErr error) {
	request := fasthttp.AcquireRequest()
	response = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(request)

	// apply functional options modifiers
	for _, option := range options {
		option(request, response)
	}

	request.Header.SetMethod("GET")
	request.SetRequestURI(uri)

	err := httpClient.Do(request, response)
	if err != nil {
		// assume no releasing request needed
		return nil, err
	}

	return response, finalErr
}

// compactPOST performs a HTTP POST request using the provided client.
// Returns the response (note: must be externally released) if request successful, else returns err != nil.
func compactPOST(
	httpClient *fasthttp.Client, uri string, data string, options ...func(*fasthttp.Request, *fasthttp.Response),
) (response *fasthttp.Response, finalErr error) {
	request := fasthttp.AcquireRequest()
	response = fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(request)

	// apply functional options modifiers
	for _, option := range options {
		option(request, response)
	}

	request.Header.SetMethod("POST")
	request.SetBodyString(data)
	request.SetRequestURI(uri)

	err := httpClient.Do(request, response)
	if err != nil {
		// assume no releasing request needed
		return nil, err
	}

	return response, finalErr
}
