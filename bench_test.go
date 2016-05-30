package grokky

import (
	"github.com/vjeantet/grok"

	"testing"
)

// RFC3339     = "2006-01-02T15:04:05Z07:00"
const testee = "2006-01-02T15:04:05Z07:00"

var global map[string]string

// find:
// tz, date, year, month, day, time, hour, min, sec
func Benchmark_logrusorgru_grokky_rfc3339(b *testing.B) {
	b.StopTimer()
	h := New()
	// from base patterns
	h.Must("YEAR", `(?:\d\d){1,2}`)
	h.Must("MONTHNUM2", `0[1-9]|1[0-2]`)
	h.Must("MONTHDAY", `(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9]`)
	h.Must("HOUR", `2[0123]|[01]?[0-9]`)
	h.Must("MINUTE", `[0-5][0-9]`)
	h.Must("SECOND", `(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?`)
	h.Must("TIMEZONE", `Z%{HOUR}:%{MINUTE}`)
	// not from base
	h.Must("DATE", "%{YEAR:year}-%{MONTHNUM2:month}-%{MONTHDAY:day}")
	h.Must("TIME", "%{HOUR:hour}:%{MINUTE:min}:%{SECOND:sec}")
	// compile the pattern
	p, err := h.Compile("%{DATE:date}T%{TIME:time}%{TIMEZONE:tz}")
	if err != nil {
		b.Fatal(err)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mss := p.Parse(testee)
		global = mss
	}
	b.ReportAllocs()
}

func berr(b *testing.B, err error) {
	if err != nil {
		b.Fatal(err)
	}
}

// find:
// tz, date, year, month, day, time, hour, min, sec
func Benchmark_vjeantet_grok_rfc3339(b *testing.B) {
	b.StopTimer()
	h, err := grok.NewWithConfig(&grok.Config{
		SkipDefaultPatterns: true,
		NamedCapturesOnly:   true,
	})
	if err != nil {
		b.Error("error creating vjeantet/grok:", err)
		b.SkipNow()
	}
	//
	berr(b, h.AddPattern("YEAR", `(?:\d\d){1,2}`))
	berr(b, h.AddPattern("MONTHNUM2", `0[1-9]|1[0-2]`))
	berr(b,
		h.AddPattern("MONTHDAY", `(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9]`))
	berr(b, h.AddPattern("HOUR", `2[0123]|[01]?[0-9]`))
	berr(b, h.AddPattern("MINUTE", `[0-5][0-9]`))
	berr(b, h.AddPattern("SECOND", `(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?`))
	berr(b, h.AddPattern("TIMEZONE", `Z%{HOUR}:%{MINUTE}`))
	//
	berr(b,
		h.AddPattern("DATE", "%{YEAR:year}-%{MONTHNUM2:month}-%{MONTHDAY:day}"))
	berr(b, h.AddPattern("TIME", "%{HOUR:hour}:%{MINUTE:min}:%{SECOND:sec}"))
	// the pattern
	berr(b,
		h.AddPattern("MAIN", "%{DATE:date}T%{TIME:time}%{TIMEZONE:tz}"))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		mss, _ := h.Parse("%{MAIN}", testee)
		global = mss
	}
	b.ReportAllocs()
}

// go test -v -run the_difference

func Test_the_difference(t *testing.T) {
	t.Logf(`show the difference between logrusorgru/grokky and vjeantet/grok
pattern '%s{NUM:one} %s{NUMBERS}'
  where NUMBERS is '%s{NUM:one} %s{NUM:two}'
    and NUM     is '\d' (single number)
Input is: '1 2 3'`, "%", "%", "%", "%") // <- go vet (stupid things happens)
	{
		h := New()
		h.Add("NUM", `\d`)
		h.Add("NUMBERS", "%{NUM:one} %{NUM:two}")
		h.Add("RES", "%{NUM:one} %{NUMBERS}")
		p, err := h.Get("RES")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("logrusorgru/grokky:", p.Parse("1 2 3"))
	}
	{
		h, err := grok.NewWithConfig(&grok.Config{
			NamedCapturesOnly:   true,
			SkipDefaultPatterns: true,
		})
		if err != nil {
			t.Fatal(err)
		}
		h.AddPattern("NUM", `\d`)
		h.AddPattern("NUMBERS", "%{NUM:one} %{NUM:two}")
		h.AddPattern("RES", "%{NUM:one} %{NUMBERS}")
		mss, err := h.Parse("%{RES}", "1 2 3")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("vjeantet/grok:", mss)
	}
}
