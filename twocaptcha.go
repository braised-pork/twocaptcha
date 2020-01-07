package twocaptcha

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/valyala/fasthttp"
)

// SettingInfo contains settings info like time between successive checking requests. These
// settings are passed into the captcha constructor by the user.
type SettingInfo struct {
	TimeBetweenRequests int
}

// Instance contains fields required for interfacing with the 2captcha API including the user's
// API key, necessary settings (time between requests) and HTTP client for sending requests.
type Instance struct {
	APIKey     string
	Settings   SettingInfo
	HTTPClient *fasthttp.Client
}

type captchaResponse struct {
	Status   int    `json:"status"` // 0 means error, 1 represents valid request
	Response string `json:"request"`
}

func checkResponse(response *fasthttp.Response) (result bool) {
	result = true

	return result
}

func NewInstance(apiKey string, settings SettingInfo) (instance Instance, finalErr error) {
OuterLoop:
	for {
		// Verify fields within Settings correctly inputted
		if settings.TimeBetweenReqs <= 0 {
			finalErr = errors.New("invalid setting TimeBetweenReqs value")
			break OuterLoop
		}

		httpClient := &fasthttp.Client{}

		var balRespStruct captchaResponse
		requestURL := capResultURL + "&action=getBalance&key=" + apiKey
		// Verify api key by checking remaining balance - don't do anything if balance empty
		for retryRequest := true; retryRequest; {
			request := fasthttp.AcquireRequest()
			request.Header.SetMethod("GET")
			request.SetRequestURI(requestURL)

			response := fasthttp.AcquireResponse()
			instance.HTTPClient.Do(request, response)
			if checkResponse(response) {
				if err := json.Unmarshal(response.Body(), &balRespStruct); err != nil {
					finalErr = errorUnmarshal
					fasthttp.ReleaseRequest(request)
					fasthttp.ReleaseResponse(response)
					break OuterLoop
				}
				retryRequest = false
			}
			fasthttp.ReleaseRequest(request)
			fasthttp.ReleaseResponse(response)
		}
		if err := containsError(&balRespStruct); err != nil {
			finalErr = err
			break OuterLoop
		}

		instance.APIKey = apiKey
		instance.Settings = settings
		instance.HTTPClient = httpClient
		break OuterLoop
	}

	return instance, finalErr
}

func (instance Instance) solveCaptcha(createTaskURL string) (solution string, finalErr error) {
OuterLoop:
	for {
		var checkSolutionURL string
		// Doing Atoi alot takes ... resources?
		// - Maybe turn SettingInfo into interface{} vs string map
		// - Remove SettingInfo and instead have each setting as a field
		timeToSleep := time.Second * time.Duration(instance.Settings.TimeBetweenReqs)

	CreateTaskLoop:
		for {
			var taskStruct captchaResponse
			for retryRequest := true; retryRequest; {
				request := fasthttp.AcquireRequest()
				request.Header.SetMethod("GET")
				request.SetRequestURI(createTaskURL)

				response := fasthttp.AcquireResponse()
				instance.HTTPClient.Do(request, response)
				if checkResponse(response) {
					if err := json.Unmarshal(response.Body(), &taskStruct); err != nil {
						finalErr = errorUnmarshal
						fasthttp.ReleaseRequest(request)
						fasthttp.ReleaseResponse(response)
						break OuterLoop
					}
					retryRequest = false
				}
				fasthttp.ReleaseRequest(request)
				fasthttp.ReleaseResponse(response)
			}

			if err := containsError(&taskStruct); err != nil {
				if err == errorNoSlot {
					time.Sleep(timeToSleep)
					continue CreateTaskLoop
				}

				finalErr = err
				break OuterLoop
			}

			captchaTaskID := taskStruct.Response // only includes task ID
			checkSolutionURL = fmt.Sprintf(
				"%s&key=%s&action=get&id=%s",
				capResultURL, instance.APIKey, captchaTaskID,
			)

			break CreateTaskLoop
		}

	SolutionLoop:
		for {
			var solutionStruct captchaResponse
			for retryRequest := true; retryRequest; {
				request := fasthttp.AcquireRequest()
				request.Header.SetMethod("GET")
				request.SetRequestURI(checkSolutionURL)

				response := fasthttp.AcquireResponse()
				instance.HTTPClient.Do(request, response)
				if checkResponse(response) {
					if err := json.Unmarshal(response.Body(), &solutionStruct); err != nil {
						finalErr = errorUnmarshal
						fasthttp.ReleaseRequest(request)
						fasthttp.ReleaseResponse(response)
						break OuterLoop
					}
					retryRequest = false
				}
				fasthttp.ReleaseRequest(request)
				fasthttp.ReleaseResponse(response)
			}
			if err := containsError(&solutionStruct); err != nil {
				if err == errorNotReady {
					time.Sleep(timeToSleep)
					continue SolutionLoop
				}

				finalErr = err
				break OuterLoop
			}

			solution = solutionStruct.Response
			break OuterLoop
		}
	}

	return solution, finalErr
}

// SolveRecaptchaV2 solves Google RecaptchaV2
func (instance *Instance) SolveRecaptchaV2(sitekey string, siteurl string) (solution string, finalErr error) {
	createTaskURL := fmt.Sprintf(
		"%s&key=%s&method=userrecaptcha&googlekey=%s&pageurl=%s",
		capRequestURL, instance.APIKey, sitekey, siteurl,
	)

	solution, finalErr = instance.solveCaptcha(createTaskURL)

	return solution, finalErr
}

// SolveRecaptchaV3 solves Google RecaptchaV3
func (instance *Instance) SolveRecaptchaV3(
	sitekey string, siteurl string, action string, minScore string,
) (solution string, finalErr error) {
OuterLoop:
	for {
		if !stringInSlice(validV3Scores, minScore) {
			finalErr = errors.New("invalid recaptchaV3 minScore (.1/.3/.9)")
			break OuterLoop
		}

		createTaskURL := fmt.Sprintf(
			"%s&key=%s&method=userrecaptcha&version=v3&googlekey=%s&pageurl=%s&action=%s&min_score=%s",
			capRequestURL, instance.APIKey, sitekey, siteurl, action, minScore,
		)

		solution, finalErr = instance.solveCaptcha(createTaskURL)
	}

	return solution, finalErr
}

// SolveFuncaptcha solves Arkose Funcaptcha
func (instance *Instance) SolveFuncaptcha(sitekey string, surl string, siteurl string) (solution string, finalErr error) {
	createTaskURL := fmt.Sprintf(
		"%s&key=%s&method=funcaptcha&publickey=%s&surl=%s&pageurl=%s",
		capRequestURL, instance.APIKey, sitekey, surl, siteurl,
	)

	solution, finalErr = instance.solveCaptcha(createTaskURL)

	return solution, finalErr
}
