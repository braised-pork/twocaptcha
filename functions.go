package twocaptcha

func containsError(responseStruct *captchaResponse) (finalErr error) {
	if responseStruct.Status == 0 {
		for key, value := range captchaErrors {
			if responseStruct.Response == key {
				finalErr = value // error
				break
			}
		}
	}

	return finalErr
}

func keyInMap(inputMap map[string]string, key string) (result bool) {
	_, result = inputMap[key]
	return result
}

func stringInSlice(inputSlice []string, key string) (result bool) {
	for _, item := range inputSlice {
		if key == item {
			result = true
			break
		}
	}

	return result
}
