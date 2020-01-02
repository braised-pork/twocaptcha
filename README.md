# twocaptcha
Golang package for solving captchas through the 2captcha API

# Install
```
go get -u "https://github.com/braised-pork/twocaptcha"
```

# Usage
```
apiKey := "insert_apikey_here"
settingParams := twocaptcha.Settings{
  TimeBetweenReqs: 5,
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

