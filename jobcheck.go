package jobcheck

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type JobChecker struct {
	timer, exempt    time.Duration // ticker, 豁免期  exempt/ticker 就是平均分成几个节点，每个节点将承载一些检查任务
	ticker           *time.Ticker
	hit              map[string]int64
	capacity         chan []string
	check            func(keys []string) error
	clean, cleanmask int // 清理hit的标记
	sync.Mutex
}

func NewJobChecker(timer, exempt time.Duration, check func(keys []string) error) *JobChecker {
	jc := &JobChecker{
		timer:     timer,
		exempt:    exempt,
		check:     check,
		hit:       make(map[string]int64, 2048),
		capacity:  make(chan []string, 4),
		ticker:    time.NewTicker(timer),
		cleanmask: 8,
	}
	go jc.Loop()

	return jc
}

func (jc *JobChecker) Capacity() chan []string {
	<-jc.ticker.C

	// delete hit
	jc.cleanHit()

	return jc.capacity
}

func (jc *JobChecker) cleanHit() {
	jc.clean++
	if jc.clean < jc.cleanmask {
		return
	}
	jc.clean = 0

	jc.Lock()
	defer jc.Unlock()

	last := time.Now().Add(-1 * jc.exempt).Unix()
	for key, pre := range jc.hit {
		if pre <= last {
			delete(jc.hit, key)
		}
	}
	fmt.Println("#- hit", len(jc.hit))
}

func (jc *JobChecker) Loop() {
	for {
		keys := <-jc.capacity
		last := time.Now().Add(-1 * jc.exempt).Unix()

		func() {
			jc.Lock()
			defer jc.Unlock()

			idx := 0
			for i, key := range keys {
				pre, ex := jc.hit[key]
				if ex && pre > last {
					continue
				}
				if idx != i {
					keys[idx] = keys[i]
				}
				idx++
			}
			if idx <= 0 {
				fmt.Println("#! check ", keys[:idx])
				return
			}
			if err := jc.check(keys[:idx]); err != nil {
				fmt.Println("cherr err:", err)
			} else {
				now := time.Now().Unix()
				for i := 0; i < idx; i++ {
					jc.hit[keys[i]] = now + int64(rand.Intn(int(jc.exempt.Seconds())))
				}
			}

		}()
	}
}
