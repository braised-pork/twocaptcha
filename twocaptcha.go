package twocaptcha

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/valyala/fasthttp"
)

// Constants, shouldn't be modified (left as var because slices un-constable)
var validTypes = []string{"recaptchaV2", "recaptchaV3", "funcaptcha"}
var validV3Scores = []string{".1", ".3", ".9"}
var capRequestURL = "https://2captcha.com/in.php?json=1"
var capResultURL = "https://2captcha.com/res.php?json=1"
var captchaErrors = map[string]error{
	// Automatically handled errors
	"CAPCHA_NOT_READY":        errors.New("handled by program"),
	"ERROR_NO_SLOT_AVAILABLE": errors.New("handled by program"),
	// API key errors (for both endpoints)
	"ERROR_WRONG_USER_KEY":     errors.New("invalidly formatted api key"),
	"ERROR_KEY_DOES_NOT_EXIST": errors.New("invalid api key"),
	// https://2captcha.com/in.php
	"ERROR_ZERO_BALANCE":          errors.New("[in] empty account balance"),
	"IP_BANNED":                   errors.New("[in] ip banned, contact 2captcha"),
	"ERROR_BAD_TOKEN_OR_PAGEURL":  errors.New("[in] recapv2 invalid token/pageurl"),
	"ERROR_GOOGLEKEY":             errors.New("[in] recapv2 invalid sitekey"),
	"MAX_USER_TURN":               errors.New("[in] too many requests, temp 10s ban"),
	"ERROR_ZERO_CAPTCHA_FILESIZE": errors.New("[in] this shouldn't happen"),
	// https://2captcha.com/res.php
	"CAPTCHA_NOT_READY":        errors.New("[res] captcha not ready"),
	"ERROR_CAPTCHA_UNSOLVABLE": errors.New("[res] unsolvable captcha"),
	"ERROR_WRONG_ID_FORMAT":    errors.New("[res] invalidly formatted captcha id"),
	"ERROR_WRONG_CAPTCHA_ID":   errors.New("[res] invalid captcha id"),
	"ERROR_BAD_DUPLICATES":     errors.New("[res] not enough matches"),
	"ERROR_EMPTY_ACTION":       errors.New("[res] action not found"),
}

// Settings contains settings info passed into the NewInstance constructor
type Settings struct {
	timeBetweenReqs int
}

// RecaptchaV2 contains fields to be passed into the NewInstance constructor
type RecaptchaV2 struct {
	sitekey string
	siteurl string
}

// RecaptchaV3 contains fields to be passed into the NewInstance constructor
type RecaptchaV3 struct {
	sitekey  string
	siteurl  string
	action   string
	minScore string
}

// Funcaptcha contains fields to be passed into the NewInstance constructor
type Funcaptcha struct {
	sitekey string
	siteurl string
	surl    string
}

// CaptchaInstance represents an individual captcha instance interfacing with the 2captcha API.
// Different combinations of captcha type and parameters (captchaInfo) require separate instances;
// for instance, even for the same website solving both RecaptchaV2 and RecaptchaV3 require two
// separate instances.
type CaptchaInstance struct {
	APIKey        string
	CaptchaType   string // must be within validTypes
	CreateTaskURL string
	// recaptchaV2 - sitekey, siteurl
	// recaptchaV3 - sitekey, siteurl, action, minScore (.1/.3/.9)
	// funcaptcha  - sitekey, surl, siteurl
	SettingInfo Settings
	// "timeBetweenReqs" int: time between checking requests
	HTTPClient *fasthttp.Client
}

type captchaResponse struct {
	Status   int    // 0 usually represents error, 1 represents valid request
	Response string `json:"request"` // response body (called request)
}

// checkResponse checks whether a request was successful - for instance, some websites send
// zero-length responses and 503s. This function primarily acts like a just-in-case and
// currently does nothing.
func checkResponse(rawResponse *fasthttp.Response) (result bool) {
	result = true
	return result
}

func checkError(responseStruct *captchaResponse) (errKey string, err error) {
	if responseStruct.Status == 0 {
		for key, value := range captchaErrors {
			if responseStruct.Response == key {
				errKey = key
				err = value // error
				break
			}
		}
	}
	return errKey, err
}

// keyInMap checks whether a given key exists within a map ([string]string)
func keyInMap(inputMap map[string]string, key string) (result bool) {
	_, result = inputMap[key]
	return result
}

// stringInSlice checks whether an input slice (of strings) contains a string
func stringInSlice(inputSlice []string, key string) (result bool) {
	for _, item := range inputSlice {
		if key == item {
			result = true
			break
		}
	}
	return result
}

// NewInstance creates and populates a new CaptchaInstance. If any error is encountered during
// initialization, NewInstance returns an empty CaptchaInstance and whatever error was found, else
// it returns the populated instance and nil error.
func NewInstance(
	apiKey string, captchaInfo interface{}, settingInfo Settings,
) (instance CaptchaInstance, finalErr error) {
OuterLoop:
	for {
		var captchaType string

		if _, ok := captchaInfo.(RecaptchaV2); ok {
			captchaType = "recaptchaV2"
		} else if _, ok := captchaInfo.(RecaptchaV3); ok {
			captchaType = "recaptchaV3"
		} else if _, ok := captchaInfo.(Funcaptcha); ok {
			captchaType = "funcaptcha"
		}

		// captchaType value unchanged from initial = invalid captchaInfo struct type
		if captchaType == "" {
			finalErr = errors.New("invalid captcha struct type")
			break OuterLoop
		}

		// Verify fields within Settings correctly inputted
		if settingInfo.timeBetweenReqs <= 0 {
			finalErr = errors.New("invalid setting timeBetweenReqs value")
			break OuterLoop
		}

		httpClient := &fasthttp.Client{}

		var balanceStruct captchaResponse
		requestURL := capResultURL + "&action=getBalance&key=" + apiKey
		// Verify api key by checking remaining balance - don't do anything if balance empty
		for retryRequest := true; retryRequest; {
			request := fasthttp.AcquireRequest()
			request.Header.SetMethod("GET")
			request.SetRequestURI(requestURL)
			response := fasthttp.AcquireResponse()
			httpClient.Do(request, response)
			if checkResponse(response) {
				if err := json.Unmarshal(response.Body(), &balanceStruct); err != nil {
					finalErr = errors.New("error unmarshalling (this shouldn't happen)")
					fasthttp.ReleaseRequest(request)
					fasthttp.ReleaseResponse(response)
					break OuterLoop
				}
				retryRequest = false
			}
			fasthttp.ReleaseRequest(request)
			fasthttp.ReleaseResponse(response)
		}

		if _, err := checkError(&balanceStruct); err != nil {
			finalErr = err
			break OuterLoop
		}

		createTaskURL := capRequestURL + "&key=" + apiKey + "&"
		// Captcha type should've already been verified earlier = type assertion should work
		switch captchaType {
		case "recaptchaV2":
			createTaskURL += "method=userrecaptcha&googlekey=" + captchaInfo.(RecaptchaV2).sitekey +
				"&pageurl=" + captchaInfo.(RecaptchaV2).siteurl
		case "recaptchaV3":
			createTaskURL += "method=userrecaptcha&version=v3&googlekey=" + captchaInfo.(RecaptchaV3).sitekey +
				"&pageurl=" + captchaInfo.(RecaptchaV3).siteurl + "&action=" + captchaInfo.(RecaptchaV3).action +
				"&min_score=" + captchaInfo.(RecaptchaV3).minScore
		case "funcaptcha":
			createTaskURL += "method=funcaptcha&publickey=" + captchaInfo.(Funcaptcha).sitekey +
				"&surl=" + captchaInfo.(Funcaptcha).surl + "&pageurl=" + captchaInfo.(Funcaptcha).siteurl
		default:
			finalErr = errors.New("invalid captcha type (this shouldn't happen!)")
			break OuterLoop
		}

		instance.APIKey = apiKey
		instance.CaptchaType = captchaType
		instance.CreateTaskURL = createTaskURL
		instance.SettingInfo = settingInfo
		instance.HTTPClient = httpClient
		break OuterLoop

	}

	return instance, finalErr
}

// SolveCaptcha solves for a given captcha type and returns the solution and error, if any.
// If any errors are encountered, SolveCaptcha returns an empty solution string and error.
func (instance *CaptchaInstance) SolveCaptcha() (solution string, finalErr error) {
OuterLoop:
	for {
		var checkSolutionURL string
		// Doing Atoi alot takes ... resources?
		// - Maybe turn SettingInfo into interface{} vs string map
		// - Remove SettingInfo and instead have each setting as a field
		timeToSleep := time.Second * time.Duration(instance.SettingInfo.timeBetweenReqs)

	CreateTaskLoop:
		for {
			var taskStruct captchaResponse
			// Create captcha solving task using instance's CreateTaskURL
			for retryRequest := true; retryRequest; {
				request := fasthttp.AcquireRequest()
				request.Header.SetMethod("GET")
				request.SetRequestURI(instance.CreateTaskURL)
				response := fasthttp.AcquireResponse()
				instance.HTTPClient.Do(request, response)
				if checkResponse(response) {
					if err := json.Unmarshal(response.Body(), &taskStruct); err != nil {
						finalErr = errors.New("error unmarshalling (this shouldn't happen)")
						fasthttp.ReleaseRequest(request)
						fasthttp.ReleaseResponse(response)
						break OuterLoop
					}
					retryRequest = false
				}
				fasthttp.ReleaseRequest(request)
				fasthttp.ReleaseResponse(response)
			}

			if errKey, err := checkError(&taskStruct); err != nil {
				if errKey == "ERROR_NO_SLOT_AVAILABLE" {
					time.Sleep(timeToSleep)
					continue
				}
				finalErr = err
				break OuterLoop
			}

			captchaTaskID := taskStruct.Response // Should only include task id
			checkSolutionURL = capResultURL + "&key=" + instance.APIKey + "&action=get&id=" + captchaTaskID
			break CreateTaskLoop
		}

	SolutionLoop:
		for {
			var solutionStruct captchaResponse
			// Check for captcha completion, else sleep and retry
			for retryRequest := true; retryRequest; {
				request := fasthttp.AcquireRequest()
				request.Header.SetMethod("GET")
				request.SetRequestURI(checkSolutionURL)
				response := fasthttp.AcquireResponse()
				instance.HTTPClient.Do(request, response)
				if checkResponse(response) {
					if err := json.Unmarshal(response.Body(), &solutionStruct); err != nil {
						finalErr = errors.New("error unmarshalling (this shouldn't happen)")
						fasthttp.ReleaseRequest(request)
						fasthttp.ReleaseResponse(response)
						break OuterLoop
					}
					retryRequest = false
				}
				fasthttp.ReleaseRequest(request)
				fasthttp.ReleaseResponse(response)
			}

			if errKey, err := checkError(&solutionStruct); err != nil {
				if errKey == "CAPCHA_NOT_READY" {
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
