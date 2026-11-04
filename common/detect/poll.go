package detect

import (
	"context"
	"fmt"
	"time"

	skyflowError "github.com/skyflowapi/skyflow-go/common/errors"
	"github.com/skyflowapi/skyflow-go/common/logger"
)

// effective returns the options with defaults applied.
func (o PollOptions) effective() (time.Duration, int) {
	wait, attempts := o.WaitTime, o.MaxAttempts
	if wait == 0 {
		wait = defaultWaitTimeSeconds * time.Second
	}
	if attempts == 0 {
		attempts = defaultMaxAttempts
	}
	return wait, attempts
}

func isTerminal(status DetectRunStatus) bool {
	return status == DetectRunStatusSuccess || status == DetectRunStatusFailed || status == DetectRunStatusUnknown
}

// poll looks the run up until it reaches a terminal status (SUCCESS, FAILED, UNKNOWN) or the
// budget runs out, and returns the last run seen. The first lookup is immediate; after that
// the pause doubles from 1s up to 32s, and the last pause is trimmed so the total never
// exceeds WaitTime. If ctx ends while waiting, the last run is returned as when the budget
// runs out, so the caller keeps the run ID.
func (c *Controller) poll(ctx context.Context, runId string, options PollOptions, fetch func(context.Context) (*GetRunResponse, *skyflowError.SkyflowError)) (*GetRunResponse, *skyflowError.SkyflowError) {
	waitTime, maxAttempts := options.effective()
	logger.Info(fmt.Sprintf(c.logMsg(logPollStarted), runId, waitTime, maxAttempts))

	var elapsed time.Duration
	delay := time.Second
	for attempts := 1; ; attempts++ {
		run, err := fetch(ctx)
		if err != nil {
			return nil, err
		}
		if isTerminal(run.Status) {
			logger.Info(fmt.Sprintf(c.logMsg(logPollTerminal), runId, run.Status, attempts))
			return run, nil
		}
		if attempts >= maxAttempts || elapsed >= waitTime {
			logger.Info(fmt.Sprintf(c.logMsg(logPollBudgetExhausted), runId, run.Status, attempts, elapsed))
			return run, nil
		}
		pause := min(delay, waitTime-elapsed)
		if !c.sleep(ctx, pause) {
			logger.Info(fmt.Sprintf(c.logMsg(logPollCancelled), runId, run.Status))
			return run, nil
		}
		elapsed += pause
		delay = min(delay*2, maxPollDelaySeconds*time.Second)
	}
}
