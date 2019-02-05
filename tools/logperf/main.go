package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sort"
	"strconv"
	"time"
)

var stdin = os.Stdin

type taskSample struct {
	line        int
	id          int
	duration    time.Duration
	description string
}

func (t taskSample) String() string {
	return fmt.Sprintf("# %d %d %s %v", t.line, t.id, t.description, t.duration)
}

type ensureSample struct {
	line        int
	duration    time.Duration
	description string
}

func (e ensureSample) String() string {
	return fmt.Sprintf("# %d %s %v", e.line, e.description, e.duration)
}

type samples struct {
	ensure []ensureSample
	task   []taskSample
}

func parse(in io.Reader) (*samples, error) {
	re := regexp.MustCompile(`PERF: (start|end) (task|ensure) ([0-9a-z"]+)( (".*")| took (.*))?`)
	scanner := bufio.NewScanner(in)
	lineNum := 0

	tasks := make(map[int]*taskSample, 100)
	ensures := make([]ensureSample, 0, 100)
	var ensure ensureSample
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		res := re.FindAllStringSubmatch(line, -1)
		if len(res) == 0 {
			continue
		}
		match := res[0]
		start := match[1] == "start"
		var duration time.Duration
		if !start {
			var err error
			duration, err = time.ParseDuration(match[6])
			if err != nil {
				log.Printf("failed to parse duration %q: %v", match[6], err)
				continue
			}
		}
		switch match[2] {
		case "task":
			id, err := strconv.Atoi(match[3])
			if err != nil {
				log.Printf("invalid task id: %v: %v", match[3], err)
			}
			if start {
				tasks[id] = &taskSample{
					line:        lineNum,
					id:          id,
					description: match[4],
				}
			} else {
				task, ok := tasks[id]
				if !ok {
					log.Printf("unmatched task sample")
					continue
				}
				task.duration = duration
			}
		case "ensure":
			if start {
				ensure = ensureSample{
					line:        lineNum,
					description: match[3],
				}
			} else {
				if ensure == (ensureSample{}) {
					log.Printf("empty ensure sample")
					continue
				}
				ensure.duration = duration
				ensures = append(ensures, ensure)
				ensure = ensureSample{}
			}
		}

		// fmt.Printf("# %d results: %q\n", lineNum, match)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	tasksList := make([]taskSample, 0, len(tasks))
	for _, task := range tasks {
		tasksList = append(tasksList, *task)
	}
	return &samples{
		task:   tasksList,
		ensure: ensures,
	}, nil
}

type byEnsureDuration []ensureSample

func (e byEnsureDuration) Len() int           { return len(e) }
func (e byEnsureDuration) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e byEnsureDuration) Less(i, j int) bool { return e[i].duration < e[j].duration }

type byTaskDuration []taskSample

func (t byTaskDuration) Len() int           { return len(t) }
func (t byTaskDuration) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t byTaskDuration) Less(i, j int) bool { return t[i].duration < t[j].duration }

func main() {
	samples, err := parse(stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	sort.Sort(byEnsureDuration(samples.ensure))
	sort.Sort(byTaskDuration(samples.task))

	fmt.Printf("### ensure\n")
	for _, sample := range samples.ensure {
		fmt.Println(sample)
	}

	fmt.Printf("### tasks\n")
	for _, sample := range samples.task {
		fmt.Println(sample)
	}
}
