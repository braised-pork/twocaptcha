# twocaptcha
Golang package for solving captchas through the 2captcha API

# Install
```
go get -u "https://github.com/braised-pork/twocaptcha"
```

# Required Parameters
No clean way to go about this, so...  
- apiKey [string]: 2captcha API key (32 characters)  
- captchaType [string]: Currently recaptchaV2/recaptchaV3/funcaptcha  
- captchaParams [map[string]string]: Info pertaining to each captcha type  
  recaptchaV2: sitekey, siteurl  
  recaptchaV3: sitekey, siteurl, action, minScore (.1/.3/.9)  
  funcaptcha: sitekey, surl (NOT siteurl), siteurl  
- settingParams [map[string]string]: Instance-specific info  
  timeBetweenReqs [string]: time between captcha status requests (seconds)  

# Usage
```
apiKey := "insert_apikey_here"
settingParams := twocaptcha.Settings{
  timeBetweenReqs: 5,
}

instance, err := twocaptcha.NewInstance(
  apiKey,
  settingParams,
}
if err != nil {
  // do something with err
}

solution, err := instance.SolveRecaptchaV2(
  sitekey: "insert_sitekey_here",
  siteurl: "insert_siteurl_here",
)
if err != nil {
  // do something with err
}

solution, err := instance.SolveCaptcha()
if err != nil {
  // do something with err
}
```

