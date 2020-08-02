package twocaptcha

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/valyala/fasthttp"
)

// parse.go contains request wrappers for evaluating API responses.
// TODO how to properly wrap errors?

// captchaPack contains status and message packaged the 2captcha JSON response.
type captchaPack struct {
	Status  int    `json:"status"` // error if Status != 1
	Message string `json:"message"`
}

// map error message to return actions
// -1: server error             |  3: authentication / ip error
//  0: no error                 |  4: parameters error / unsolvable
//  1: captcha not ready        |  5: no money in your account lol
//  2: timeout, idk what to do
var errorActions = map[string]int{
	"CAPCHA_NOT_READY":               1,
	"ERROR_WRONG_USER_KEY":           3,
	"ERROR_KEY_DOES_NOT_EXIST":       3,
	"ERROR_IP_NOT_ALLOWED":           3,
	"IP_BANNED":                      3,
	"ERROR_BAD_PROXY":                3,
	"ERROR_ZERO_CAPTCHA_FILESIZE":    4,
	"ERROR_TOO_BIG_CAPTCHA_FILESIZE": 4,
	"ERROR_WRONG_FILE_EXTENSION":     4,
	"ERROR_IMAGE_TYPE_NOT_SUPPORTED": 4,
	"ERROR_UPLOAD":                   4,
	"ERROR_BAD_TOKEN_OR_PAGEURL":     4,
	"ERROR_GOOGLEKEY":                4,
	"ERROR_CAPTCHAIMAGE_BLOCKED":     4,
	"ERROR_BAD_PARAMETERS":           4,
	"ERROR_CAPTCHA_UNSOLVABLE":       4,
	"ERROR_WRONG_ID_FORMAT":          4,
	"ERROR_WRONG_CAPTCHA_ID":         4,
	"ERROR_BAD_DUPLICATES":           4,
	"ERROR_EMPTY_ACTION":             4,
	"ERROR_ZERO_BALANCE":             5,
}

// captchaWrap wraps API requests by parsing errors, messages, etc. after parsing to JSON.
// Returns the message and action (for further action) if successfully parsed, else returns err != nil
func (solver *Solver) captchaWrap(response *fasthttp.Response) (message string, action int, finalErr error) {
	// unmarshal response into JSON
	var pack captchaPack
	if err := json.Unmarshal(response.Body(), &pack); err != nil {
		return "", -1, err
	}

	// return normally if no error encountered
	if pack.Status == 1 {
		return pack.Message, 0, nil
	}

	// evaluate timeout-specific errors
	var timeoutDuration time.Duration
	switch pack.Message {
	case "MAX_USER_TURN": // too many requests
		timeoutDuration = 10 * time.Second
	case "ERROR_NO_SLOT_AVAILABLE": // no solving slot
		timeoutDuration = 5 * time.Second
	default:
		if strings.HasPrefix(pack.Message, "ERROR: ") { // other rate limiting error codes
			errorCode, _ := strconv.Atoi(strings.Split(pack.Message, " ")[1])

			switch errorCode {
			case 1: // nothing wrong, return
				return pack.Message, 0, nil
			// errors, refer to https://2captcha.com/2captcha-api#limits
			case 1003: // 30 second timeout
				timeoutDuration = 30 * time.Second
			case 1002, 1005: // 5 minute timeout
				timeoutDuration = 5 * time.Minute
			case 1001, 1004: // 10 minute timeout
				timeoutDuration = 10 * time.Minute
			}
		}
	}

	// overwrite global timeout variable
	now := time.Now()
	if timeoutDuration != time.Duration(0) && now.Add(timeoutDuration).After(solver.timeout) { // TODO double-check
		solver.timeout = now.Add(timeoutDuration) // overwrite timeout with new timeout
		return pack.Message, 2, nil
	}

	var ok bool // check for "missed" errors, just in case
	action, ok = errorActions[pack.Message]
	if !ok {
		panic(errors.New("unknown error " + pack.Message))
	}

	return pack.Message, action, nil // if errored action, caller wraps err with message
}
