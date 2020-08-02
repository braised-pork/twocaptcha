package twocaptcha

import (
	"time"

	"github.com/pkg/errors"
	"github.com/remeh/sizedwaitgroup"
	"github.com/valyala/fasthttp"
)

// TODO find better way to manage errors, e.g.: panic cancel threshold

var dummyError = errors.New("")                       // non captcha-related error
var inEndpoint = "https://2captcha.com/in.php?json=1" // ... maintainability?
var resEndpoint = "https://2captcha.com/res.php?json=1"
var checkTimeout = 100 * time.Millisecond // time to wait between attempting to close solutions channel

// waitTimeout waits out the duration of any timeout if any exists
func (solver *Solver) waitTimeout() {
	now := time.Now()
	if now.Before(solver.timeout) { // wait out the timeout
		time.Sleep(solver.timeout.Sub(now))
	}
}

// solveCaptcha solves a single captcha through 2captcha using the provided URL
// Any request-related errors are returned and treated as insignificant, TODO flesh out
func (solver *Solver) solveCaptcha(client *fasthttp.Client) (result string, finalErr error) {
	var response *fasthttp.Response
	var err error

	// create task, return if error encountered (redo)
	solver.waitTimeout() // check timeout before request
	response, err = compactGET(client, solver.taskURL)
	if err != nil {
		return "", dummyError
	}

	message, action, err := solver.captchaWrap(response)
	fasthttp.ReleaseResponse(response)
	switch action {
	case 0: // no error
		break
	case 2: // timeout error (handle specially?)
		return "", errors.New(message)
	case 3, 4, 5: // regular errors
		return "", errors.New(message)
	}
	solver.numTasks++ // increment before request for accurate counter, TODO maybe just increment at end

	resultURL := resEndpoint + "?key=" + solver.apiKey + "&action=get&id=" + message // Sprintf performance >:(
ResultLoop:
	for { // keep trying to get solution
		// check solution status
		solver.waitTimeout() // check timeout before request
		response, err = compactGET(client, resultURL)
		if err != nil {
			solver.numTasks-- // remember to decrement on error!
			return "", dummyError
		}

		message, action, err = solver.captchaWrap(response)
		fasthttp.ReleaseResponse(response)
		switch action {
		case 0: // captcha solved
			return message, nil
		case 1: // waiting on captcha, wait and retry
			time.Sleep(5 * time.Second)
			continue ResultLoop
		case 2: // timeout error (handle specially?) - shouldn't occur
			solver.numTasks-- // remember to decrement on error!
			return "", errors.New(message)
		case 3, 4, 5: // regular errors
			solver.numTasks-- // remember to decrement on failure!
			return "", errors.New(message)
		}
	}
}

// persistCaptcha represents a goroutine (thread) that continuously solves captchas.
func (solver *Solver) persistCaptcha(swg *sizedwaitgroup.SizedWaitGroup) {
	defer swg.Done()

	var result string
	var err error
	client := &fasthttp.Client{} // TODO add customizable client (proxies?)
	for {                        // persist solving continuously
		select { // solve until cancelled
		case <-solver.ctx.Done():
			return
		default:
			break
		}

		// check if total number of tasks created, terminate thread if so
		// if task fails, that thread will keep retrying so it's ok to do this
		if solver.solveType == 1 && solver.numTasks >= int(float64(solver.TotalCaptchas)*solver.Multiplier) {
			return
		}

		// solve captcha, pass error to channel if exists
		// note that only significant errors are passed, i.e.: not slot issues
		result, err = solver.solveCaptcha(client)
		if err != nil && err != dummyError {
			solver.Channels.Errors <- err
			continue
		}

		solver.Channels.Solved <- result
		solver.numSolved++ // don't care about atomic
		if solver.solveType == 1 && solver.numSolved >= solver.TotalCaptchas {
			solver.cancel() // channels closed elsewhere?
		}
	}
}

// solvingRuntime manages the threads which solve captchas and closes channels if necessary.
// TODO maybe return analytics from the function?
func (solver *Solver) SolvingRuntime(taskURL string) {
	swg := sizedwaitgroup.New(solver.Threads)
	for i := 0; i < solver.Threads; i++ {
		swg.Add()
		solver.persistCaptcha(&swg)
	}

	// close solutions channel if first solving type, ONCE all solutions consumed
	if solver.solveType == 1 {
		// keep waiting until solutions consumed... TODO find more elgegant way than for loop
		for {
			if len(solver.Channels.Solved) == 0 {
				close(solver.Channels.Solved)
				break
			}

			time.Sleep(checkTimeout)
		}
	}
}
