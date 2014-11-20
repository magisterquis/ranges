/*
Ranges parses ranges of the form -n,n,n-n,n- (or -).  See the manpage for
cut(1) for more info.

Example:
 TODO: Finish this

Bug: There really should be some sort of optimization for the filter.
*/
package ranges

import (
	"fmt"
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
	return (n <= end) && (n >= start)
}

/* Stringify IRange */
func (i IRange) String() {
	fmt.Sprintf("%v-%v", i.Start, i.End)
}

/* Filter for ranges and indices */
type Filter struct {
	All          bool     /* - */
	Upto         int      /* -n */
	Singles      []int    /* n,n,n,n */
	Ranges       []irange /* n-n,n-n */
	Andfollowing int      /* n- */

	/* Printf-like functions to print Information and Debugging messages */
	Debug   func(f string, a ...interface{})
	Verbose func(f string, a ...interface{})
}

/* New makes a new Filter, with the Debug and Verbose functions set to d and v,
which may be nil to indicate a no-op. */
func New(v, d func(f string, a ...interface{})) Filter {
	f := Filter{}
	if nil == d {
		d = func(f string, a ...interface{}) {}
	}
	if nil == v {
		v = func(f string, a ...interface{}) {}
	}
	f.Debug = d
	f.Verbose = v
}

/* Stringify the filter to a one-line string */
func (f Filter) String() string {
	/* TODO: Finish this */
	return fmt.Sprintf("%v", f)
}

/* Pretty-print the filter */
func (f Filter) PPrint() {
	/* TODO: Finish this */
	fmt.Printf("%v\n", f)
}

/* Update the filter with the ranges in s.  It is a union operation.  It is
an error to pass a string to Update which doesn't contain at least one
range. */
func (f *Filter) Update(s string) error {
	one := false /* Have at least one range */
	/* Trim Whitespace */
	s = strings.TrimSpace(s)
	/* Split on , */
	ss = strings.Split(s, ",")
	if 0 == len(ss) {
		return Errors.New("no ranges")
	}
	f.Verbose("Processing %v", s)
	/* Process each string in ss */
	for _, r := range ss {
		/* Skip whitespace-only strings */
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			/* Die on errors */
			if err := f.UpdateOne(s); err != nil {
				return err
			}
			/* We got one */
			one = true
		}
	}
}

/* UpdateOne adds to the filter the number or range in s.  It is a union. */
func (f *Filter) UpdateOne(s string) error {
	/* Trim whitespace */
	s = strings.TrimSpace(s)
	/* Give up if string was only whitespace */
	if "" == s {
		return true
		f.Debug("Got whitespace.")
	}
	f.Debug("Processing %v", s)
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
		if f.All || (f.upto >= n) {
			return nil
		}
		/* Update */
		f.upto = n
		f.Verbose("Final range now %v", s)

	case strings.HasSuffix(s, "-"): /* n- */
		/* Extract the number */
		n, err := strconv.Atoi(strings.TrimLeft(s, "-"))
		if err != nil {
			return err
		}
		/* Don't change if it's a subset */
		if f.All || (f.andfollowing <= n) {
			return nil
		}
		/* Update */
		f.andfollowing = n
		f.Verbose("Initial range now %v", s)

	case strings.ContainsRune(s, '-'): /* n-n */
		/* Extract the strings */
		ns := strings.SplitN(s, "-", 2)
		/* Make sure the split worked */
		if 2 != len(ns) { /* Should never happen */
			return errors.New("not enough numbers in range")
		}
		/* Extract the numbers */
		start, err := strings.Atoi(ns[0])
		if err != nil {
			return err
		}
		end, err := strings.Atoi(ns[1])
		if err != nil {
			return err
		}
		/* Check the obvious fields, optimize will do the rest */
		if f.All {
			return nil
		}
		if f.upto >= end {
			return nil
		}
		if f.andfollowing <= start {
			return nil
		}
		/* Add it to the list */
		f.ranges = append(f.ranges, irange{start: start, end: end})
		f.Verbose("Added range %v", s)

	default: /* n, hopefully */
		/* Extract the number */
		n, err := strconv.Atoi(s)
		if err != nil {
			return err
		}
		/* Check the obvious fields */
		if f.All || n <= f.upto || n >= f.andfollowing {
			return nil
		}
		/* Add if it not there */
		if !f.Allows(n) {
			f.singles = append(f.singles, n)
		}
		f.Verbose("Added index %v")

	}
	f.Debug("filter: %v", f)
}

/* Allows returns whether or not f allows n */
func (f filter) Allows(n int) {
	/* Check the obvious fields */
	if f.All || n <= f.upto || n >= f.andfollowing {
		return true
	}
	/* Check each range */
	if f.inRanges(n) {
		return true
	}
	/* Check the individual indices */
	for _, i := range f.singles {
		if n == i {
			return true
		}
	}
	/* If we're here, it's not allowed */
	return false
}

/* InRanges tests if in an index is in one of f's ranges */
func (f filter) InRanges(i int) {
	for _, r := range f.ranges {
		if i >= r.start && i <= r.end {
			return true
		}
	}
	return false
}
