// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2019 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"

	"github.com/snapcore/snapd/client"
	"github.com/snapcore/snapd/i18n"
)

var shortPerformanceHelp = i18n.G("Print performance metrics")
var longPerformanceHelp = i18n.G(`
The performance command prints various performance metrics of the system
`)

type cmdPerformance struct {
	clientMixin
	timeMixin
	ChangeID string `long:"change"`
}

func init() {
	addDebugCommand("performance", shortPerformanceHelp, longPerformanceHelp, func() flags.Commander {
		return &cmdPerformance{}
	}, nil, nil)
}

func (c *cmdPerformance) Execute(args []string) error {
	if c.ChangeID != "" {
		return c.showChange(c.ChangeID)
	}
	return nil
}

func printSample(w io.Writer, level int, prefix string, s *client.TaskPerformanceSample) {
	fmt.Fprintf(w, "%s\t%s\t%s\n", prefix, s.Duration.Round(time.Millisecond), strings.Repeat("  ", level)+s.Summary)
	for _, c := range s.Samples {
		printSample(w, level+1, prefix+"-", c)
	}
}

func (c *cmdPerformance) showChange(chid string) error {
	chg, err := queryChange(c.client, chid)
	if err != nil {
		return err
	}

	w := tabWriter()

	var totalActiveTime time.Duration
	level := 0
	for _, t := range chg.Tasks {
		summary := t.Summary
		if t.Status == "Doing" && t.Progress.Total > 1 {
			summary = fmt.Sprintf("%s (%.2f%%)", summary, float64(t.Progress.Done)/float64(t.Progress.Total)*100.0)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", t.Status, t.ActiveTime.Round(time.Millisecond), summary)
		for _, s := range t.PerformanceSamples {
			printSample(w, level, "-", s)
		}
		totalActiveTime += t.ActiveTime
	}

	fmt.Fprintf(w, "id: %v\n", chg.ID)
	fmt.Fprintf(w, "kind: %v\n", chg.Kind)
	fmt.Fprintf(w, "total run time: %v\n", totalActiveTime)

	w.Flush()

	return nil

}
