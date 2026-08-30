package fn

import (
	"strings"
	"unicode/utf8"
)

// RFC 3492 punycode 参数。
const (
	_pcBase        = 36
	_pcTmin        = 1
	_pcTmax        = 26
	_pcSkew        = 38
	_pcDamp        = 700
	_pcInitialBias = 72
	_pcInitialN    = 0x80
	_pcDelimiter   = '-'
	_pcPrefix      = "xn--"
)

// asciiDomain 把国际化域名转成 punycode（琳.tw → xn--6q5d.tw）。
// 已是 ASCII 的域名原样返回；已经是 xn-- 的标签不重复编码。
// 这里只做编码，不实现 IDNA 的映射与校验规则，对书签场景足够。
func asciiDomain(host string) string {
	if isASCII(host) {
		return host
	}
	labels := strings.Split(strings.TrimSuffix(host, "."), ".")
	for i, label := range labels {
		if label == "" || isASCII(label) || strings.HasPrefix(strings.ToLower(label), _pcPrefix) {
			continue
		}
		labels[i] = _pcPrefix + punycodeEncode(strings.ToLower(label))
	}
	out := strings.Join(labels, ".")
	if strings.HasSuffix(host, ".") {
		out += "."
	}
	return out
}

// punycodeEncode 实现 RFC 3492 的编码部分。
func punycodeEncode(s string) string {
	var out []byte
	b, h := 0, 0
	for _, r := range s {
		if r < _pcInitialN {
			out = append(out, byte(r))
			b++
		}
	}
	if b > 0 {
		out = append(out, _pcDelimiter)
	}
	// RFC 3492：h 必须从基本码点数量起步，否则主循环永远走不到 total。
	h = b

	n := rune(_pcInitialN)
	delta := 0
	bias := _pcInitialBias
	total := utf8.RuneCountInString(s)

	for h < total {
		// 取下一个待编码的最小码点
		m := rune(0x10FFFF)
		for _, r := range s {
			if r >= n && r < m {
				m = r
			}
		}

		delta += int(m-n) * (h + 1)
		n = m

		for _, r := range s {
			if r < n {
				delta++
			}
			if r == n {
				q := delta
				for k := _pcBase; ; k += _pcBase {
					t := k - bias
					if t < _pcTmin {
						t = _pcTmin
					} else if t > _pcTmax {
						t = _pcTmax
					}
					if q < t {
						break
					}
					out = append(out, pcDigitToBasic(t+(q-t)%(_pcBase-t)))
					// RFC 3492 的推广变长整数：必须先减去 t 再整除。
					q = (q - t) / (_pcBase - t)
				}
				out = append(out, pcDigitToBasic(q))
				bias = pcAdapt(delta, h+1, h == b)
				delta = 0
				h++
			}
		}
		delta++
		n++
	}
	return string(out)
}

func pcAdapt(delta, numPoints int, firstTime bool) int {
	if firstTime {
		delta /= _pcDamp
	} else {
		delta /= 2
	}
	delta += delta / numPoints

	k := 0
	for delta > ((_pcBase-_pcTmin)*_pcTmax)/2 {
		delta /= _pcBase - _pcTmin
		k += _pcBase
	}
	return k + ((_pcBase-_pcTmin+1)*delta)/(delta+_pcSkew)
}

// pcDigitToBasic 按 RFC 3492 公式 d + 22 + 75*(d<26) 生成输出字符。
func pcDigitToBasic(d int) byte {
	if d < 26 {
		return byte(d + 97) // 'a' + d
	}
	return byte(d + 22) // '0' + (d - 26)
}
