//
// Copyright (c) 2016-2017 Konstanin Ivanov <kostyarin.ivanov@gmail.com>.
// All rights reserved. This program is free software. It comes without
// any warranty, to the extent permitted by applicable law. You can
// redistribute it and/or modify it under the terms of the Do What
// The Fuck You Want To Public License, Version 2, as published by
// Sam Hocevar. See LICENSE file for more details or see below.
//

//
//        DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
//                    Version 2, December 2004
//
// Copyright (C) 2004 Sam Hocevar <sam@hocevar.net>
//
// Everyone is permitted to copy and distribute verbatim or modified
// copies of this license document, and changing it is allowed as long
// as the name is changed.
//
//            DO WHAT THE FUCK YOU WANT TO PUBLIC LICENSE
//   TERMS AND CONDITIONS FOR COPYING, DISTRIBUTION AND MODIFICATION
//
//  0. You just DO WHAT THE FUCK YOU WANT TO.
//

package grokky

import (
	"github.com/vjeantet/grok"

	"testing"
)

// Intel Core i5-6200U
// DDR4 8G 2133 MHz
// Linux Ubuntu 17.04 4.12.4-041204-generic
//
// go test -bench . -benchtime=1m
//
// Benchmark_logrusorgru_grokky_rfc3339-4    30000000    3201 ns/op  1297 B/op  5 allocs/op
// Benchmark_vjeantet_grok_rfc3339-4         30000000    2967 ns/op  1329 B/op  5 allocs/op
// Benchmark_grokkyVsGrokApacheLog/grokky-4    200000  447480 ns/op  5098 B/op  6 allocs/op
// Benchmark_grokkyVsGrokApacheLog/grok-4      200000  473763 ns/op  5609 B/op  6 allocs/op
// PASS
// ok      github.com/logrusorgru/grokky   385.076s

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
		b.Skip("error creating vjeantet/grok:", err)
	}

	for _, np := range []struct{ name, pattern string }{
		{"YEAR", `(?:\d\d){1,2}`},
		{"MONTHNUM2", `0[1-9]|1[0-2]`},
		{"MONTHDAY", `(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9]`},
		{"HOUR", `2[0123]|[01]?[0-9]`},
		{"MINUTE", `[0-5][0-9]`},
		{"SECOND", `(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?`},
		{"TIMEZONE", `Z%{HOUR}:%{MINUTE}`},
		//
		{"DATE", "%{YEAR:year}-%{MONTHNUM2:month}-%{MONTHDAY:day}"},
		{"TIME", "%{HOUR:hour}:%{MINUTE:min}:%{SECOND:sec}"},
		// the pattern
		{"MAIN", "%{DATE:date}T%{TIME:time}%{TIMEZONE:tz}"},
	} {
		if err = h.AddPattern(np.name, np.pattern); err != nil {
			b.Skip("error adding pattern:", err)
		}
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		if global, err = h.Parse("%{MAIN}", testee); err != nil {
			b.Skip("parsing error:", err)
		}
	}
	b.ReportAllocs()
}

// go test -v -run the_difference

func Test_the_difference(t *testing.T) {

	t.Log(`show the difference between logrusorgru/grokky and vjeantet/grok
pattern '%{NUM:one} %{NUMBERS}'
  where NUMBERS is '%{NUM:one} %{NUM:two}'
    and NUM     is '\d' (single number)
Input is: '1 2 3'`)

	const input = "1 2 3"

	t.Run("logrusorgru/grokky", func(t *testing.T) {
		h := New()
		h.Add("NUM", `\d`)
		h.Add("NUMBERS", "%{NUM:one} %{NUM:two}")
		h.Add("RES", "%{NUM:one} %{NUMBERS}")
		p, err := h.Get("RES")
		if err != nil {
			t.Fatal(err)
		}
		t.Log("result is:", p.Parse(input))
	})

	t.Run("vjeantet/grok", func(t *testing.T) {
		h, err := grok.NewWithConfig(&grok.Config{
			NamedCapturesOnly:   true,
			SkipDefaultPatterns: true,
		})
		if err != nil {
			t.Skip("error:", err)
		}
		h.AddPattern("NUM", `\d`)
		h.AddPattern("NUMBERS", "%{NUM:one} %{NUM:two}")
		h.AddPattern("RES", "%{NUM:one} %{NUMBERS}")
		mss, err := h.Parse("%{RES}", input)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("result is:", mss)
	})

	// Input is "1 2 3", output of this test:
	//
	// --- PASS: Test_the_difference/logrusorgru/grokky (0.00s)
	//     bench_test.go:147: result is: map[one:1 two:3]
	// --- PASS: Test_the_difference/vjeantet/grok (0.00s)
	//     bench_test.go:165: result is: map[one:2 two:3]
	//

	// E.g. for pattern %{one:\d %{one: \d, two: \d}},
	// grokky matches 1 and 3 (closer to top, sicne the
	// second "one" is nested), and the grok returns
	// 2 and 3 (closer to tail), and nesting level doesn't
	// matter.

}

type testParseFunc func(string, string) (map[string]string, error)

var testGrokkyParse = func() testParseFunc {

	type Pair struct {
		Name    string
		Pattern string
	}

	patterns := []Pair{
		{"LOGLEVEL", `([Aa]lert|ALERT|[Tt]race|TRACE|[Dd]ebug|DEBUG|[Nn]otice|NOTICE|[Ii]nfo|INFO|[Ww]arn?(?:ing)?|WARN?(?:ING)?|[Ee]rr?(?:or)?|ERR?(?:OR)?|[Cc]rit?(?:ical)?|CRIT?(?:ICAL)?|[Ff]atal|FATAL|[Ss]evere|SEVERE|EMERG(?:ENCY)?|[Ee]merg(?:ency)?)`},
		{"USERNAME", `[a-zA-Z0-9._-]+`},
		{"HOSTNAME", `\b(?:[0-9A-Za-z][0-9A-Za-z-]{0,62})(?:\.(?:[0-9A-Za-z][0-9A-Za-z-]{0,62}))*(\.?|\b)`},
		{"USER", `%{USERNAME}`},
		{"EMAILLOCALPART", `[a-zA-Z][a-zA-Z0-9_.+-=:]+`},
		{"EMAILADDRESS", `%{EMAILLOCALPART}@%{HOSTNAME}`},
		{"HTTPDUSER", `%{EMAILADDRESS}|%{USER}`},
		{"INT", `(?:[+-]?(?:[0-9]+))`},
		{"BASE10NUM", `([+-]?(?:[0-9]+(?:\.[0-9]+)?)|\.[0-9]+)`},
		{"NUMBER", `(?:%{BASE10NUM})`},
		{"BASE16NUM", `(0[xX]?[0-9a-fA-F]+)`},
		{"POSINT", `\b(?:[1-9][0-9]*)\b`},
		{"NONNEGINT", `\b(?:[0-9]+)\b`},
		{"WORD", `\b\w+\b`},
		{"NOTSPACE", `\S+`},
		{"SPACE", `\s*`},
		{"DATA", `.*?`},
		{"GREEDYDATA", `.*`},
		{"QUOTEDSTRING", `"([^"\\]*(\\.[^"\\]*)*)"|\'([^\'\\]*(\\.[^\'\\]*)*)\'`},
		{"UUID", `[A-Fa-f0-9]{8}-(?:[A-Fa-f0-9]{4}-){3}[A-Fa-f0-9]{12}`},
		{"CISCOMAC", `(?:(?:[A-Fa-f0-9]{4}\.){2}[A-Fa-f0-9]{4})`},
		{"WINDOWSMAC", `(?:(?:[A-Fa-f0-9]{2}-){5}[A-Fa-f0-9]{2})`},
		{"COMMONMAC", `(?:(?:[A-Fa-f0-9]{2}:){5}[A-Fa-f0-9]{2})`},
		{"MAC", `(?:%{CISCOMAC}|%{WINDOWSMAC}|%{COMMONMAC})`},
		{"IPV6", `((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?`},
		{"IPV4", `(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)`},
		{"IP", `(?:%{IPV6}|%{IPV4})`},
		{"HOST", `%{HOSTNAME}`},
		{"IPORHOST", `(?:%{IP}|%{HOSTNAME})`},
		{"HOSTPORT", `%{IPORHOST}:%{POSINT}`},
		{"UNIXPATH", `(/[\w_%!$@:.,-]?/?)(\S+)?`},
		{"WINPATH", `([A-Za-z]:|\\)(?:\\[^\\?*]*)+`},
		{"PATH", `(?:%{UNIXPATH}|%{WINPATH})`},
		{"TTY", `(?:/dev/(pts|tty([pq])?)(\w+)?/?(?:[0-9]+))`},
		{"URIPROTO", `[A-Za-z]+(\+[A-Za-z+]+)?`},
		{"URIHOST", `%{IPORHOST}(?::%{POSINT:port})?`},
		{"URIPATH", `(?:/[A-Za-z0-9$.+!*'(){},~:;=@#%_\-]*)+`},
		{"URIPARAM", `\?[A-Za-z0-9$.+!*'|(){},~@#%&/=:;_?\-\[\]<>]*`},
		{"URIPATHPARAM", `%{URIPATH}(?:%{URIPARAM})?`},
		{"URI", `%{URIPROTO}://(?:%{USER}(?::[^@]*)?@)?(?:%{URIHOST})?(?:%{URIPATHPARAM})?`},
		{"MONTH", `\b(?:Jan(?:uary|uar)?|Feb(?:ruary|ruar)?|M(?:a|ä)?r(?:ch|z)?|Apr(?:il)?|Ma(?:y|i)?|Jun(?:e|i)?|Jul(?:y)?|Aug(?:ust)?|Sep(?:tember)?|O(?:c|k)?t(?:ober)?|Nov(?:ember)?|De(?:c|z)(?:ember)?)\b`},
		{"MONTHNUM", `(?:0?[1-9]|1[0-2])`},
		{"MONTHNUM2", `(?:0[1-9]|1[0-2])`},
		{"MONTHDAY", `(?:(?:0[1-9])|(?:[12][0-9])|(?:3[01])|[1-9])`},
		{"DAY", `(?:Mon(?:day)?|Tue(?:sday)?|Wed(?:nesday)?|Thu(?:rsday)?|Fri(?:day)?|Sat(?:urday)?|Sun(?:day)?)`},
		{"YEAR", `(\d\d){1,2}`},
		{"HOUR", `(?:2[0123]|[01]?[0-9])`},
		{"MINUTE", `(?:[0-5][0-9])`},
		{"SECOND", `(?:(?:[0-5]?[0-9]|60)(?:[:.,][0-9]+)?)`},
		{"TIME", `([^0-9]?)%{HOUR}:%{MINUTE}(?::%{SECOND})([^0-9]?)`},
		{"DATE_US", `%{MONTHNUM}[/-]%{MONTHDAY}[/-]%{YEAR}`},
		{"DATE_EU", `%{MONTHDAY}[./-]%{MONTHNUM}[./-]%{YEAR}`},
		{"ISO8601_TIMEZONE", `(?:Z|[+-]%{HOUR}(?::?%{MINUTE}))`},
		{"ISO8601_SECOND", `(?:%{SECOND}|60)`},
		{"TIMESTAMP_ISO8601", `%{YEAR}-%{MONTHNUM}-%{MONTHDAY}[T ]%{HOUR}:?%{MINUTE}(?::?%{SECOND})?%{ISO8601_TIMEZONE}?`},
		{"DATE", `%{DATE_US}|%{DATE_EU}`},
		{"DATESTAMP", `%{DATE}[- ]%{TIME}`},
		{"TZ", `(?:[PMCE][SD]T|UTC)`},
		{"DATESTAMP_RFC822", `%{DAY} %{MONTH} %{MONTHDAY} %{YEAR} %{TIME} %{TZ}`},
		{"DATESTAMP_RFC2822", `%{DAY}, %{MONTHDAY} %{MONTH} %{YEAR} %{TIME} %{ISO8601_TIMEZONE}`},
		{"DATESTAMP_OTHER", `%{DAY} %{MONTH} %{MONTHDAY} %{TIME} %{TZ} %{YEAR}`},
		{"DATESTAMP_EVENTLOG", `%{YEAR}%{MONTHNUM2}%{MONTHDAY}%{HOUR}%{MINUTE}%{SECOND}`},
		{"HTTPDERROR_DATE", `%{DAY} %{MONTH} %{MONTHDAY} %{TIME} %{YEAR}`},
		{"SYSLOGTIMESTAMP", `%{MONTH} +%{MONTHDAY} %{TIME}`},
		{"PROG", `[\x21-\x5a\x5c\x5e-\x7e]+`},
		{"SYSLOGPROG", `%{PROG:program}(?:\[%{POSINT:pid}\])?`},
		{"SYSLOGHOST", `%{IPORHOST}`},
		{"SYSLOGFACILITY", `<%{NONNEGINT:facility}.%{NONNEGINT:priority}>`},
		{"HTTPDATE", `%{MONTHDAY}/%{MONTH}/%{YEAR}:%{TIME} %{INT}`},
		{"QS", `%{QUOTEDSTRING}`},
		{"SYSLOGBASE", `%{SYSLOGTIMESTAMP:timestamp} (?:%{SYSLOGFACILITY} )?%{SYSLOGHOST:logsource} %{SYSLOGPROG}:`},
		{"COMMONAPACHELOG", `%{IPORHOST:clientip} %{HTTPDUSER:ident} %{USER:auth} \[%{HTTPDATE:timestamp}\] "(?:%{WORD:verb} %{NOTSPACE:request}(?: HTTP/%{NUMBER:httpversion})?|%{DATA:rawrequest})" %{NUMBER:response} (?:%{NUMBER:bytes}|-)`},
		{"COMBINEDAPACHELOG", `%{COMMONAPACHELOG} %{QS:referrer} %{QS:agent}`},
		{"HTTPD20_ERRORLOG", `\[%{HTTPDERROR_DATE:timestamp}\] \[%{LOGLEVEL:loglevel}\] (?:\[client %{IPORHOST:clientip}\] ){0,1}%{GREEDYDATA:errormsg}`},
		{"HTTPD24_ERRORLOG", `\[%{HTTPDERROR_DATE:timestamp}\] \[%{WORD:module}:%{LOGLEVEL:loglevel}\] \[pid %{POSINT:pid}:tid %{NUMBER:tid}\]( \(%{POSINT:proxy_errorcode}\)%{DATA:proxy_errormessage}:)?( \[client %{IPORHOST:client}:%{POSINT:clientport}\])? %{DATA:errorcode}: %{GREEDYDATA:message}`},
		{"HTTPD_ERRORLOG", `%{HTTPD20_ERRORLOG}|%{HTTPD24_ERRORLOG}`},
	}

	h := New()
	for _, p := range patterns {
		h.Add(p.Name, p.Pattern)
	}

	p, _ := h.Get("COMBINEDAPACHELOG")
	return func(name string, input string) (m map[string]string, err error) {
		m = p.Parse(input)
		return
	}

}()

var testGrokParse = func() testParseFunc {
	g, _ := grok.NewWithConfig(&grok.Config{
		NamedCapturesOnly: true,
	})
	return g.Parse
}()

func Benchmark_grokkyVsGrokApacheLog(b *testing.B) {

	for _, np := range []struct {
		name      string
		parseFunc testParseFunc
	}{
		{"grokky", testGrokkyParse},
		{"grok", testGrokParse},
	} {

		b.Run(np.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m, err := np.parseFunc("%{COMBINEDAPACHELOG}", `127.0.0.1 - - [02/Aug/2017:22:58:13 +0800] "GET / HTTP/1.1" 200 612 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:52.0) Gecko/20100101 Firefox/52.0" "-"`)
				if len(m) == 0 {
					b.Fatal(err)
				}
				m, err = np.parseFunc("%{COMBINEDAPACHELOG}", `....`)
				if len(m) != 0 {
					b.Fatal(err, m)
				}
			}

			b.ReportAllocs()
		})

	}

}

// one cycle is 3 log lines
// ------------------------
//
// $ go test -bench simpleNginxAccessLog -benchtime=1m
// Benchmark_simpleNginxAccessLog-4  3000000  26505 ns/op  1872 B/op  12 allocs/op
// PASS
// ok      github.com/logrusorgru/grokky   107.686s

// so, a second is 1*1000*1000*1000, thus 26505ns/op is 37728op/s,
// where op is 3 lines (~ 113184op/s)

// after #3
// Benchmark_simpleNginxAccessLog-4  10000000  9895 ns/op  1370 B/op  5 allocs/op

// where op is _one_ line

func Benchmark_simpleNginxAccessLog(b *testing.B) {

	// https://play.golang.org/p/XKtY84Uicf

	// using simple regular expression
	// (?P<remote_addr>[^\s]+)\s\-\s(?P<remote_user>[^\s]*)\s*\-\s\[(?P<time_local>.+)\]\s\"(?P<request>[^"]*)\"\s(?P<status>[^\s]+)\s(?P<body_bytes_sent>[^\s]+)\s\"(?P<htp_referer>[^"]+)\"\s\"(?P<user_agent>[^"]+)\"

	b.StopTimer()

	h := New()

	h.Must("NSS", `[^\s]*`) // not a space *
	h.Must("NS", `[^\s]+`)  // not a space +
	h.Must("NLB", `[^\]]+`) // not a left bracket +
	h.Must("NQS", `[^"]*`)  // not a quote *
	h.Must("NQ", `[^"]+`)   // not a double quotes +
	h.Must("A", `.*`)       // all
	h.Must("nginx", `%{NS:clientip}\s%{NSS:ident}\s%{NSS:auth}`+
		`\s\[`+
		`%{NLB:timestamp}\]\s\"`+
		`%{NS:verb}\s`+
		`%{NSS:request}\s`+
		`HTTP/%{NS:httpversion}\"\s`+
		`%{NS:response}\s`+
		`%{NS:bytes}\s\"`+
		`%{NQ:referrer}\"\s\"`+
		`%{NQ:agent}\"`+
		`%{A:blob}`)

	nginx, err := h.Get("nginx")
	if err != nil {
		b.Fatal(err)
	}

	lines := []string{
		`66.249.65.159 - - [06/Nov/2014:19:10:38 +0600] ` +
			`"GET /news/53f8d72920ba2744fe873ebc.html HTTP/1.1" ` +
			`404 177 "-" ` +
			`"Mozilla/5.0 (iPhone; CPU iPhone OS 6_0 like Mac OS X)` +
			` AppleWebKit/536.26 (KHTML, like Gecko) Version/6.0` +
			` Mobile/10A5376e Safari/8536.25 (compatible; Googlebot/2.1;` +
			` +http://www.google.com/bot.html)"`,
		`66.249.65.3 - - [06/Nov/2014:19:11:24 +0600] ` +
			`"GET /?q=%E0%A6%AB%E0%A6%BE%E0%A7%9F%E0%A6%BE%E0%A6%B0 ` +
			`HTTP/1.1" ` +
			`200 4223 "-" ` +
			`"Mozilla/5.0 (compatible; Googlebot/2.1; ` +
			`+http://www.google.com/bot.html)"`,
		`66.249.65.62 - - [06/Nov/2014:19:12:14 +0600] ` +
			`"GET /?q=%E0%A6%A6%E0%A7%8B%E0%A7%9F%E0%A6%BE HTTP/1.1" ` +
			`200 4356 "-" ` +
			`"Mozilla/5.0 (compatible; Googlebot/2.1; ` +
			`+http://www.google.com/bot.html)"`,
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		global = nginx.Parse(lines[i%len(lines)])
		if len(global) != 12 {
			b.Fatal(global)
		}
	}

	b.ReportAllocs()
}
