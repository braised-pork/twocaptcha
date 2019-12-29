package twocaptcha

import (
	"errors"

	"github.com/valyala/fasthttp"
)

// Constants, shouldn't be modified (left as var because slices un-constable)
var validTypes = []string{"recaptchaV2", "recaptchaV3", "funcaptcha"}
var capRequestURL = "https://2captcha.com/in.php"
var capResultURL = "https://2captcha.com/res.php"

// CaptchaInstance represents an individual captcha instance interfacing with the 2captcha API.
// Different combinations of captcha type and parameters (captchaInfo) require separate instances;
// for instance, even for the same website solving both RecaptchaV2 and RecaptchaV3 require two
// separate instances.
type CaptchaInstance struct {
	APIKey      string
	CaptchaType string // must be within validTypes
	CaptchaInfo map[string]interface{}
	// CaptchaType = "recaptchaV2"
	//   "sitekey":  recaptcha sitekey
	//   "siteurl":  recaptcha website url
	// CaptchaType = "recaptchaV3"
	//   "sitekey":  recaptcha sitekey
	//   "siteurl":  recaptcha website url
	//   "action":   recaptchaV3 action (get, set, etc.)
	//   "minScore": recaptchaV3 required score (.1/.3/.9)
	// CaptchaType = "funcaptcha"
	//   "key":      funcaptcha key
	//   "surl":     funcaptcha surl (NOT siteurl url)
	//   "siteurl":  funcaptcha website url
	SettingInfo map[string]interface{}
	// "timeBetweenReqs" int: time between checking requests
	HTTPClient *fasthttp.Client
}

// keyInMap checks whether a given key exists within a map ([string]string)
func keyInMap(inputMap map[string]interface{}, key string) (result bool) {
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
	apiKey string, captchaType string, captchaParams map[string]interface{}, settingParams map[string]interface{},
) (instance CaptchaInstance, finalErr error) {
OuterLoop:
	for {
		// Verify that initialization key(s) (timeBetweenReqs) exist within map (settingParams).
		if !(keyInMap(settingParams, "timeBetweenReqs")) {
			finalErr = errors.New("missing parameter(s) within settingParams")
			break OuterLoop
		}

		// Verify that passed captchaType within valid types (validTypes) for proper initialization.
		if !stringInSlice(validTypes, captchaType) {
			finalErr = errors.New("invalid captcha type")
			break
		}

		// Verify that captcha-specific keys exist within map (captchaParams), then pass entire
		// captchaParams map into instance after switch statement completes.
		switch captchaType {
		case "recaptchaV2":
			if !(keyInMap(captchaParams, "sitekey") && keyInMap(captchaParams, "siteurl")) {
				finalErr = errors.New("missing parameter(s) within captchaParams for recaptchav2")
				break OuterLoop
			}
		case "recaptchaV3":
			if !(keyInMap(captchaParams, "sitekey") && keyInMap(captchaParams, "siteurl") &&
				keyInMap(captchaParams, "action") && keyInMap(captchaParams, "minscore")) {
				finalErr = errors.New("missing parameter(s) within captchaParams for recaptchav3")
				break OuterLoop
			}
		case "funcaptcha":
			if !(keyInMap(captchaParams, "key") && keyInMap(captchaParams, "surl") &&
				keyInMap(captchaParams, "siteurl")) {
				finalErr = errors.New("missing parameter(s) within captchaParams for funcaptcha")
				break OuterLoop
			}
		default: // shouldn't happen because captchaType previously verified
			finalErr = errors.New("invalid captcha type (this shouldn't happen)")
			break OuterLoop
		}

	}

	return instance, finalErr
}
