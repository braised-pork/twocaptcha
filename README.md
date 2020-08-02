# twocaptcha
Package for serving captcha solutions from the 2captcha API.  
**Supported Captcha Types:** RecaptchaV2, RecaptchaV3, Funcaptcha, hCaptcha  

**Fixed Solver:** Solve a fixed number of captchas with an optional task multiplier for avoiding hanging tasks.  
**Persistent Solver:** Keep solving captchas using the designated number of threads until channels are closed.  
**\* Note for persistent solving: closing either the solutions or errors channel cancels the solving runtime.**  
**\* General Note: mostly untested because of schoolwork, please create issues with any errors or whatnot that occur!**

# Install
```
go get -u "https://github.com/braised-pork/twocaptcha"
```

# Usage (Fixed)
```
// Initialize the solving instance (api key, threads, total, multiplier, enable errors channel)
fixedSolver, ok := twocaptcha.NewFixedSolver("insert_api_key_here", 100, 1000, 1.1, false)
if !ok {
  // provided API key is invalid, do something
}

// Set the type of captcha to solve, with parameters
fixedSolver.SetRecaptchaV2("insert_sitekey_here", "insert_siteurl_here")

go fixedSolver.SolvingRuntime() // begin solving routine

for solution := range fixedSolver.Channels.Solved {
  // do something with captcha solution
}
```

# Usage (Persistent)
```
// Initialize the solving instance (api key, threads, enable errors channel)
persistentSolver, ok := twocaptcha.NewPersistentSolver("insert_api_key_here", 100, false)
if !ok {
  // provided API key is invalid, do something
}

// Set the type of captcha to solve, with parameters
persistentSolver.SetRecaptchaV2("insert_sitekey_here", "insert_siteurl_here")

go persistentSolver.SolvingRuntime() // begin solving routine

for solution := range persistentSolver.Channels.Solved {
  // do something with captcha solution
}

// [in an alternate goroutine ...]

close(persistentSolver.Channels.Solved) // close when you've had enough
```

