package twocaptcha

import (
	"context"
	"time"
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
func defaultSolver(apiKey string, threads int) (solver *Solver) {
	ctx, cancel := context.WithCancel(context.Background())
	return &Solver{
		Channels: Channels{
			// initialize default channels
			Solved: make(chan string, 100),
			Errors: make(chan error, 100),
		},
		Threads: threads,
		apiKey:  apiKey,
		timeout: time.Now(), // just in case?
		cancel:  cancel,
		ctx:     ctx,
	}
}

// NewFixedSolver returns a Solver instance which solves a fixed number of captchas of a given type/
func NewFixedSolver(apiKey string, threads int, totalCaptchas int, multiplier float64) (solver *Solver) {
	solver = defaultSolver(apiKey, threads)
	solver.solveType = 1
	solver.TotalCaptchas = totalCaptchas
	solver.Multiplier = multiplier

	return solver
}

func NewPersistentSolver(apiKey string, threads int) (solver *Solver) {
	solver = defaultSolver(apiKey, threads)
	solver.solveType = 2

	return solver
}

// URLRecaptchaV2 returns the URL for solving the designated RecaptchaV2
func (solver *Solver) URLRecaptchaV2(siteKey string, siteURL string) (taskURL string) {
	return inEndpoint + "&key=" + solver.apiKey + "&method=userrecaptcha&googlekey=" + siteKey + "&pageurl=" +
		siteURL // Sprintf performance >:(
}

// URLRecaptchaV3 returns the URL for solving the designated RecaptchaV3
func (solver *Solver) URLRecaptchaV3(siteKey string, siteURL string, action string, minScore string) (taskURL string) {
	return inEndpoint + "&key=" + solver.apiKey + "&method=userrecaptcha&version=v3&googlekey=" + siteKey +
		"&pageurl=" + siteURL + "&action=" + action + "&min_score=" + minScore // Sprintf performance >:(
}

// URLFuncaptcha returns the URL for solving the designated Funcaptcha
func (solver *Solver) URLFuncaptcha(siteKey string, sURL string, siteURL string) (taskURL string) {
	return inEndpoint + "&key=" + solver.apiKey + "&method=funcaptcha&publickey=" + siteKey + "&surl=" + sURL +
		"&pageurl=" + siteURL // Sprintf performance >:(
}

// URLhCaptcha returns the URL for solving the designated hCaptcha
func (solver *Solver) URLhCaptcha(siteKey string, siteURL string) (taskURL string) {
	return inEndpoint + "&key=" + solver.apiKey + "&method=hcaptcha&sitekey=" + siteKey + "&pageurl=" +
		siteURL // Sprintf performance >:(
}
