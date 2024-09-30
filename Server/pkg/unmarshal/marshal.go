package unmarshal

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"go-web-app/models/crond"
	"strings"
	"time"
)

func UnPackJob(value []byte) (ret *crond.Job, err error) {

	var job *crond.Job
	job = &crond.Job{}
	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, crond.JobDir)
}
func ExtractKillerName(jobKey string) string {
	return strings.TrimPrefix(jobKey, crond.JobKill)
}

type JobEvent struct {
	EventType int
	Job       *crond.Job
}

func BUildJobEvent(evenType int, job *crond.Job) (jobEvent *JobEvent) {
	return &JobEvent{
		EventType: evenType,
		Job:       job,
	}
}

func BuildJobSchedulePlan(job *crond.Job) (jobSchedulePlan *crond.JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	jobSchedulePlan = &crond.JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}
