/*
Ranges parses ranges of the form -n,n,n-n,n- (or -).  See the manpage for
cut(1) for more info.

Example:
 TODO: Finish this

Bug: There really should be some sort of optimization for the filter.
*/
package ranges

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

/* ranges.go
 * library to parse range strings of the form -n,n,n-n,n- or -
 * by J. Stuart McMurray
 * Created 20141119
 * Last modified 20141119
 *
 * Copyright (c) 2014 J. Stuart McMurray <kd5pbo@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

/* Irange represents a range of input */
type IRange struct {
	Start int
	End   int
}

/* Has returns whether or not n is in the range specified by i */
func (i IRange) Has(n int) bool {
	return (n <= i.End) && (n >= i.Start)
}

/* Stringify IRange */
func (i IRange) String() string {
	return fmt.Sprintf("%v-%v", i.Start, i.End)
}

/* Filter for ranges and indices */
type Filter struct {
	All              bool     /* - */
	Upto             int      /* -n */
	UptoSpec         bool     /* If Upto has been specified */
	Singles          []int    /* n,n,n,n */
	Ranges           []IRange /* n-n,n-n */
	Andfollowing     int      /* n- */
	AndfollowingSpec bool     /* If Andfollowing has been specified */

	/* Printf-like functions to print Information and Debugging messages */
	Debug   func(f string, a ...interface{})
	Verbose func(f string, a ...interface{})
}

/* New makes a new Filter, with the Debug and Verbose functions set to d and v,
which may be nil to indicate a no-op. */
func New(v, d func(f string, a ...interface{})) Filter {
	/* Make a filter */
	f := Filter{}
	/* Make noops if needed */
	if nil == d {
		d = func(f string, a ...interface{}) {}
	}
	if nil == v {
		v = func(f string, a ...interface{}) {}
	}
	/* Set output functions */
	f.Debug = d
	f.Verbose = v

	return f
}

/* Stringify the filter to a one-line string */
func (f Filter) String() string {
	/* Simpler true/false */
	a := "F" /* All */
	if f.All {
		a = "T"
	}
	u := "F" /* UptoSpec */
	if f.UptoSpec {
		u = "T"
	}
	n := "F" /* AndfollowingSpec */
	if f.AndfollowingSpec {
		n = "T"
	}
	/* TODO: Finish this */
	return fmt.Sprintf("[-: %v][-n: %v(%v)][n- %v(%v)][n-n %v]"+
		"[n %v]",
		a,
		f.Upto, u,
		f.Andfollowing, n,
		f.Ranges,
		f.Singles)
}

/* Pretty-print the filter */
func (f Filter) PPrint() {
	/* TODO: Finish this */
	fmt.Printf("%v\n", f)
}

/* Update the filter with s, which contains a comma-separated list of ranges,
or indices.  It is a union operation.  It is an error to pass a string to
Update which doesn't contain at least one range. */
func (f *Filter) Update(s string) error {
	one := false /* Have at least one range */
	/* Trim Whitespace */
	s = strings.TrimSpace(s)
	/* Split on , */
	ss := strings.Split(s, ",")
	if 0 == len(ss) {
		return errors.New("no ranges")
	}
	f.Verbose("Processing range(s): %v", s)
	/* Process each string in ss */
	for _, r := range ss {
		/* Skip whitespace-only strings */
		r = strings.TrimSpace(r)
		if len(s) > 0 {
			/* Die on errors */
			if err := f.UpdateOne(r); err != nil {
				return errors.New(fmt.Sprintf("Error "+
					"processing %v: %v", r, err))
			}
			/* We got one */
			one = true
		}
	}
	/* If we didn't even get one update, throw an error */
	if !one {
		return errors.New("no ranges or indices found")
	}
	return nil
}

/* UpdateOne adds to the filter the number or range in s.  It is a union. */
func (f *Filter) UpdateOne(s string) error {
	/* Trim whitespace */
	s = strings.TrimSpace(s)
	/* Give up if string was only whitespace */
	if "" == s {
		f.Debug("Got whitespace.")
		return nil
	}
	f.Debug("Processing range: %v", s)
	/* Work out what sort of specification: -, -n, n-, n-n, n */
	switch {
	case "-" == s: /* - */
		f.All = true
		f.Verbose("Entire input selected")

	case strings.HasPrefix(s, "-"): /* -n */
		/* Extract the number */
		n, err := strconv.Atoi(strings.TrimLeft(s, "-"))
		if err != nil {
			return err
		}
		/* Don't change if it's a subset */
		if f.All || (f.UptoSpec && f.Upto >= n) {
			return nil
		}
		/* Update */
		f.Upto = n
		f.UptoSpec = true
		f.Verbose("Initial range now %v", s)

	case strings.HasSuffix(s, "-"): /* n- */
		/* Extract the number */
		n, err := strconv.Atoi(strings.TrimRight(s, "-"))
		if err != nil {
			return err
		}
		/* Don't change if it's a subset */
		if f.All || (f.AndfollowingSpec && f.Andfollowing <= n) {
			return nil
		}
		/* Update */
		f.Andfollowing = n
		f.AndfollowingSpec = true
		f.Verbose("Final range now %v", s)

	case strings.ContainsRune(s, '-'): /* n-n */
		/* Extract the strings */
		ns := strings.SplitN(s, "-", 2)
		/* Make sure the split worked */
		if 2 != len(ns) { /* Should never happen */
			return errors.New("not enough numbers in range")
		}
		/* Extract the numbers */
		start, err := strconv.Atoi(ns[0])
		if err != nil {
			return err
		}
		end, err := strconv.Atoi(ns[1])
		if err != nil {
			return err
		}
		/* Check the obvious fields, optimize will do the rest */
		if f.All {
			return nil
		}
		if f.UptoSpec && f.Upto >= end {
			return nil
		}
		if f.AndfollowingSpec && f.Andfollowing <= start {
			return nil
		}
		/* Add it to the list */
		f.Ranges = append(f.Ranges, IRange{Start: start, End: end})
		f.Verbose("Added range %v", s)

	default: /* n, hopefully */
		/* Extract the number */
		n, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		/* Check the obvious fields */
		if f.All || (f.UptoSpec && n <= f.Upto) ||
			(f.AndfollowingSpec && n >= f.Andfollowing) {
			return nil
		}
		/* Add if it not there */
		if !f.Allows(n) {
			f.Singles = append(f.Singles, n)
		}
		f.Verbose("Added index %v")

	}
	f.Debug("filter: %v", f)
	return nil
}

/* AllowsOut returns whether or not f allows in, and an additional int that
describes which part of the filter allowed n. */
const (
	/* Return values for filter.AllowsOut */
	InRange  = iota /* In a range */
	IsIndex         /* Matched a single index */
	Above           /* Above or equal to Andfollowing */
	Below           /* Below or equal to Upto */
	AllMatch        /* All is set */
)

func (f Filter) AllowsOut(n int) (bool, int) {
	/* Check the obvious fields */
	switch {
	case f.All:
		return true, AllMatch
	case f.UptoSpec && n <= f.Upto:
		return true, Below
	case f.AndfollowingSpec && n >= f.Andfollowing:
		return true, Above
	case f.InRanges(n):
		return true, InRange
	}
	/* Check the individual indices */
	for _, i := range f.Singles {
		if n == i {
			return true, IsIndex
		}
	}
	/* If we're here, it's not allowed */
	return false, 0
}

/* Allows returns whether or not f allows n */
func (f Filter) Allows(n int) bool {
	a, _ := f.AllowsOut(n)
	return a
}

/* InRanges tests if in an index is in one of f's ranges */
func (f Filter) InRanges(i int) bool {
	for _, r := range f.Ranges {
		if i >= r.Start && i <= r.End {
			return true
		}
	}
	return false
}
