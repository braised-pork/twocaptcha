package twocaptcha

import "errors"

var validTypes = []string{"recaptchaV2", "recaptchaV3", "funcaptcha"}
var validV3Scores = []string{".1", ".3", ".9"}

const (
	capRequestURL = "https://2captcha.com/in.php?json=1"
	capResultURL  = "https://2captcha.com/res.php?json=1"
)

var ( // Error return messages (from 2captcha)
	errorNotReady    = errors.New("handled by program")
	errorNoSlot      = errors.New("handled by program")
	errorWrongKey    = errors.New("invalidly formatted api key")
	errorKeyExist    = errors.New("invalid api key")
	errorZeroBal     = errors.New("[in] empty account balance")
	errorIPBan       = errors.New("[in] IP ban, contact 2captcha")
	errorBadParam    = errors.New("[in] recaptcha invalid token/pageurl")
	errorSitekey     = errors.New("[in] recaptcha invalid sitekey")
	errorManyReq     = errors.New("[in] too many requests, temp 10s ban")
	errorZeroSize    = errors.New("[in] zero captcha filesize")
	errorUnsolvable  = errors.New("[res] captcha unsolvable")
	errorIDFormat    = errors.New("[res] invalidly formatted captcha ID")
	errorWrongID     = errors.New("[res] invalid captcha ID")
	errorBadDupe     = errors.New("[res] not enough matches")
	errorEmptyAction = errors.New("[res] action not found")
)

var ( // Error return messages (from program)
	errorUnmarshal = errors.New("error unmarshalling (shouldn't happen)")
	errorV3Score   = errors.New("invalid recaptchaV3 minScore (.1/.3/.9)")
)

var captchaErrors = map[string]error{
	// Automatically handled errors
	"CAPCHA_NOT_READY":        errorNotReady,
	"ERROR_NO_SLOT_AVAILABLE": errorNoSlot,
	// API key errors (for both endpoints)
	"ERROR_WRONG_USER_KEY":     errorWrongKey,
	"ERROR_KEY_DOES_NOT_EXIST": errorKeyExist,
	// https://2captcha.com/in.php
	"ERROR_ZERO_BALANCE":          errorZeroBal,
	"IP_BANNED":                   errorIPBan,
	"ERROR_BAD_TOKEN_OR_PAGEURL":  errorBadParam,
	"ERROR_GOOGLEKEY":             errorSitekey,
	"MAX_USER_TURN":               errorManyReq,
	"ERROR_ZERO_CAPTCHA_FILESIZE": errorZeroSize,
	// https://2captcha.com/res.php
	"ERROR_CAPTCHA_UNSOLVABLE": errorUnsolvable,
	"ERROR_WRONG_ID_FORMAT":    errorIDFormat,
	"ERROR_WRONG_CAPTCHA_ID":   errorWrongID,
	"ERROR_BAD_DUPLICATES":     errorBadDupe,
	"ERROR_EMPTY_ACTION":       errorEmptyAction,
}
