# twocaptcha
Golang package for solving captchas (currently RecaptchaV2, RecaptchaV3, and Funcaptcha) through the 2captcha API.

# Install
```
go get -u github.com/braised-pork/twocaptcha/
```

# Required Parameters For Each Captcha Type
No idea how to organize this, but...
RecaptchaV2 - sitekey, siteurl
RecaptchaV3 - sitekey, siteurl, action, minScore (.1/.3/.9)
Funcaptcha - sitekey, surl (NOT siteurl), siteurl

# Required Settings (Unfortunately Go doesn't support optional methods)
timeBetweenReqs (string) - Cooldown between checking captcha status in seconds (as string). Recommended time is 5 seconds (parameter "5").
 
# Usage
First, create a new captcha instance. Instances should be unique to each captcha type and website - for instance, solving captchas for a website utilizing both RecaptchaV2 and RecaptchaV3 would require two instances, one for managing each captcha type.
```
apiKey = "insert_apikey_here"
captchaType = "recaptchaV2/recaptchaV3/funcaptcha"
// Captcha parameter example for RecaptchaV2
captchaParams := map[string]string{
  "sitekey": "insert_sitekey_here",
  "siteurl": "insert_siteurl_here",
}
settingParams := map[string]string{
  "timeBetweenReqs": "insert_time_here",
}

instance, err := twocaptcha.NewInstance(
  apiKey,
  captchaType,
  captchaParams,
  settingparams,
}
if err != nil {
  // Do something with err
}

solution, err := instance.SolveCaptcha()
if err != nil {
  // Do something with err
}
```
