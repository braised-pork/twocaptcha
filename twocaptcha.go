package twocaptcha

import (
	"context"
	"strings"
	"time"

	"github.com/valyala/fasthttp"
)

// Advantage for multiplier: reduces overall solve time due to hanging
// Different methods for persistent solving:
// 1) X threads for Y * Z total captchas with float multiplier Z, cancel after finished
// 2) X threads continuously solving captchas, externally cancelled

// TODO find more elegant way to create multiple solving instances from one config?
// TODO break into different chunks for mix-and-match
// Solver stores parameters and statistics for the current session
type Solver struct {
	// should I un-export fields?
	Channels
	Threads       int     // 1/2) solving threads
	TotalCaptchas int     // 1) captchas desired
	Multiplier    float64 // 1) solving multiplier, init at 1!
	// internal usage
	errEnable bool               // whether to use errors channel
	solveType int                // type 1 or 2 solving
	numTasks  int                // 1) tasks counter
	numSolved int                // 1) solved counter
	apiKey    string             // api key for authentication
	taskURL   string             // for creating captchas
	timeout   time.Time          // global timeout
	cancel    context.CancelFunc // for terminating solving
	ctx       context.Context    // ^^
}

type Channels struct {
	Solved chan string // captcha solutions
	Errors chan error  // any encountered errors
}

// defaultSolver returns a default Solver instance to be modified.
// Also checks whether the API key is valid and returns ok == false if not so.
func defaultSolver(apiKey string, threads int, errEnable bool) (solver *Solver, ok bool) {
	balanceURL := resEndpoint + "&key=" + apiKey + "&action=getBalance"
	response, _ := compactGET(&fasthttp.Client{}, balanceURL) // ignore error?
	validKey := !strings.Contains(string(response.Body()), "ERROR")
	fasthttp.ReleaseResponse(response)

	if validKey {
		ctx, cancel := context.WithCancel(context.Background())
		solver := &Solver{
			Channels: Channels{
				// initialize default channels
				Solved: make(chan string, 100),
			},
			Threads: threads,
			apiKey:  apiKey,
			timeout: time.Now(), // just in case?
			cancel:  cancel,
			ctx:     ctx,
		}
		if errEnable {
			solver.Channels.Errors = make(chan error, 100)
		}

		return solver, true
	} else {
		return nil, false
	}
}

// NewFixedSolver returns a Solver instance which solves a fixed number of captchas of a given type.
func NewFixedSolver(apiKey string, threads int, totalCaptchas int, multiplier float64, errEnable bool) (solver *Solver, ok bool) {
	solver, ok = defaultSolver(apiKey, threads, errEnable)
	if ok {
		if errEnable {
			solver.errEnable = true
		}
		solver.solveType = 1
		solver.TotalCaptchas = totalCaptchas
		solver.Multiplier = multiplier
	}

	return solver, ok
}

// NewPersistentSolver returns a Solver instance which continuously solves captchas until the output channel is closed.
func NewPersistentSolver(apiKey string, threads int, errEnable bool) (solver *Solver, ok bool) {
	solver, ok = defaultSolver(apiKey, threads, errEnable)
	if ok {
		if errEnable {
			solver.errEnable = true
		}
		solver.solveType = 2
	}

	return solver, ok
}

// SetRecaptchaV2 sets the URL for solving the designated RecaptchaV2.
func (solver *Solver) SetRecaptchaV2(siteKey string, siteURL string) {
	solver.taskURL = inEndpoint + "&key=" + solver.apiKey + "&method=userrecaptcha&googlekey=" + siteKey + "&pageurl=" +
		siteURL // Sprintf performance >:(
}

// SetRecaptchaV3 sets the URL for solving the designated RecaptchaV3.
func (solver *Solver) SetRecaptchaV3(siteKey string, siteURL string, action string, minScore string) {
	solver.taskURL = inEndpoint + "&key=" + solver.apiKey + "&method=userrecaptcha&version=v3&googlekey=" + siteKey +
		"&pageurl=" + siteURL + "&action=" + action + "&min_score=" + minScore // Sprintf performance >:(
}

// SetFuncaptcha sets the URL for solving the designated Funcaptcha.
func (solver *Solver) SetFuncaptcha(siteKey string, sURL string, siteURL string) {
	solver.taskURL = inEndpoint + "&key=" + solver.apiKey + "&method=funcaptcha&publickey=" + siteKey + "&surl=" + sURL +
		"&pageurl=" + siteURL // Sprintf performance >:(
}

// SethCaptcha sets the URL for solving the designated hCaptcha.
func (solver *Solver) SethCaptcha(siteKey string, siteURL string) {
	solver.taskURL = inEndpoint + "&key=" + solver.apiKey + "&method=hcaptcha&sitekey=" + siteKey + "&pageurl=" +
		siteURL // Sprintf performance >:(
}
