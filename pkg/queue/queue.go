package queue

import (
	"sync"
	"time"

	"github.com/google/btree"
	"github.com/thatique/kuade/pkg/uuid"
)

var (
	// Loggerf can be used to override the default logging destination. Such
	// log messages in this library should be logged at info or higher.
	Loggerf = func(format string, args ...interface{}) {}

	defaultUUID = uuid.Generate()
)

// Default number of btree degrees
const btreeDegrees = 64

type Job interface {
	// for logging or stats
	GetName() string

	// Fire the job
	Fire() error
}

type delayedJob struct {
	id    uuid.UUID
	job   Job
	delay time.Time
}

func (dj *delayedJob) isExpired() bool {
	return time.Now().After(dj.delay)
}

func (dj *delayedJob) Less(item btree.Item) bool {
	dj2 := item.(*delayedJob)
	if dj2.delay.After(dj.delay) {
		return true
	}
	if dj.delay.After(dj2.delay) {
		return false
	}
	return dj.id.String() < dj2.id.String()
}

type Queue struct {
	mu       sync.RWMutex
	jobCh    chan Job
	exps     *btree.BTree
	dsp      *dispatcher
	quitChan chan bool

	// function to execute when delayed job need to run. returning false make this
	// job not deleted
	onExpiredJob func(jobName string, now time.Time) bool
}

func NewQueue(maxWorkers, maxQueueSize int) *Queue {
	jobCh := make(chan Job, maxQueueSize)
	exps := btree.New(btreeDegrees)

	dsp := newDispatcher(jobCh, maxWorkers)
	dsp.run()

	queue := &Queue{jobCh: jobCh, exps: exps, dsp: dsp, quitChan: make(chan bool)}
	go queue.backgroundManager()

	return queue
}

func (q *Queue) Stop() {
	q.dsp.stop()
	q.quitChan <- true
}

func (q *Queue) SetOnExpiredJob(fn func(jobName string, now time.Time) bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.onExpiredJob = fn
}

func (q *Queue) Len() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.exps.Len()+len(q.jobCh)
}

func (q *Queue) Push(job Job) {
	q.jobCh <- job
}

func (q *Queue) Later(expired time.Duration, job Job) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.exps.ReplaceOrInsert(&delayedJob{
		id:    uuid.Generate(),
		delay: time.Now().Add(expired),
		job:   job,
	})
}

func (q *Queue) getExpiredJobs() []*delayedJob {
	q.mu.RLock()
	defer q.mu.RUnlock()
	var expired []*delayedJob
	q.exps.AscendLessThan(&delayedJob{
		id:    defaultUUID,
		delay: time.Now(),
	}, func(item btree.Item) bool {
		expired = append(expired, item.(*delayedJob))
		return true
	})

	return expired
}

func (q *Queue) processExpiredJobs(jobs []*delayedJob) <-chan bool {
	var wg sync.WaitGroup
	out := make(chan bool)

	send := func(itm *delayedJob) {
		q.mu.Lock()
		defer q.mu.Unlock()
		if q.onExpiredJob == nil {
			q.jobCh <- itm.job

			q.exps.Delete(itm)
		} else {
			needRun := q.onExpiredJob(itm.job.GetName(), time.Now().UTC())
			if needRun {
				q.jobCh <- itm.job
				q.exps.Delete(itm)
			}
		}
		wg.Done()
	}

	wg.Add(len(jobs))
	for _, job := range jobs {
		go send(job)
	}

	go func() {
		wg.Wait()
		out <- true
	}()

	return out
}

func (q *Queue) backgroundManager() {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-q.quitChan:
			return
		case <-t.C:
			d := q.processExpiredJobs(q.getExpiredJobs())
			<-d
		}
	}
}

type worker struct {
	id         int
	jobQueue   chan Job
	workerPool chan chan Job
	quitChan   chan bool
}

func newWorker(id int, workerPool chan chan Job) *worker {
	return &worker{
		id:         id,
		jobQueue:   make(chan Job),
		workerPool: workerPool,
		quitChan:   make(chan bool),
	}
}

func (w *worker) start() {
	go func() {
		for {
			w.workerPool <- w.jobQueue

			select {
			case job := <-w.jobQueue:
				Loggerf("worker%d: started %s job", w.id, job.GetName())
				err := job.Fire()
				if err != nil {
					Loggerf("worker%d: %s job failed with %s", w.id, job.GetName(), err.Error())
				} else {
					Loggerf("worker%d: %s job completed", w.id, job.GetName())
				}

			case <-w.quitChan:
				Loggerf("worker%d quitting", w.id)
				return
			}
		}
	}()
}

func (w *worker) stop() {
	w.quitChan <- true
}

type dispatcher struct {
	workerPool chan chan Job
	maxWorkers int
	workers    []*worker
	jobQueue   chan Job
	quitChan   chan bool
}

func newDispatcher(jobQueue chan Job, maxWorkers int) *dispatcher {
	workerPool := make(chan chan Job, maxWorkers)

	return &dispatcher{
		jobQueue:   jobQueue,
		maxWorkers: maxWorkers,
		workers:    make([]*worker, maxWorkers),
		workerPool: workerPool,
		quitChan:   make(chan bool),
	}
}

func (d *dispatcher) run() {
	for i := 0; i < d.maxWorkers; i++ {
		w := newWorker(i+1, d.workerPool)
		d.workers[i] = w
		w.start()
	}

	go d.dispatch()
}

func (d *dispatcher) stop() {
	for _, w := range d.workers {
		w.stop()
	}
	d.quitChan <- true
}

func (d *dispatcher) dispatch() {
	for {
		select {
		case job := <-d.jobQueue:
			go func() {
				Loggerf("fetching workerJobQueue for: %s\n", job.GetName())
				workerJobQueue := <-d.workerPool
				Loggerf("adding %s to workerJobQueue\n", job.GetName())
				workerJobQueue <- job
			}()
		case <-d.quitChan:
			return
		}
	}
}
