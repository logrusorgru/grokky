//
// Copyright (c) 2016 Konstanin Ivanov <kostyarin.ivanov@gmail.com>.
// All rights reserved. This program is free software. It comes without
// any warranty, to the extent permitted by applicable law. You can
// redistribute it and/or modify it under the terms of the Do What
// The Fuck You Want To Public License, Version 2, as published by
// Sam Hocevar. See LICENSE.md file for more details or see below.
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

// RFC3339     = "2006-01-02T15:04:05Z07:00"
const testee = "2006-01-02T15:04:05Z07:00"

var global map[string]string

type _parse func(string, string) (map[string]string, error)

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

var _grokkyParse = func() _parse {
	type Pair struct {
		Key   string
		Value string
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
		h.Add(p.Key, p.Value)
	}
	p, _ := h.Get("COMBINEDAPACHELOG")
	return func(name string, input string) (m map[string]string, err error) {
		m = p.Parse(input)
		return
	}
}()

var _grokParse = func() _parse {
	g, _ := grok.NewWithConfig(&grok.Config{
		NamedCapturesOnly: true,
	})
	return g.Parse
}()

func BenchmarkGrokkyVsGrokApacheLog(b *testing.B) {
	for n, p := range map[string]_parse{"grokky": _grokkyParse, "grok": _grokParse} {
		b.Run(n, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				m, err := p("%{COMBINEDAPACHELOG}", `127.0.0.1 - - [02/Aug/2017:22:58:13 +0800] "GET / HTTP/1.1" 200 612 "-" "Mozilla/5.0 (Macintosh; Intel Mac OS X 10.12; rv:52.0) Gecko/20100101 Firefox/52.0" "-"`)
				if len(m) == 0 {
					b.Fatal(err)
				}
				m, err = p("%{COMBINEDAPACHELOG}", `....`)
				if len(m) != 0 {
					b.Fatal(err, m)
				}
			}
		})
	}
}
