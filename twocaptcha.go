package twocaptcha

import (
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
