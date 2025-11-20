// utils/food_similarity.go

package utils

import (
	"math"
	"strings"
	"unicode"
)

// 한글 자모 상수
const (
	SBase  = 0xAC00
	LCount = 19
	VCount = 21
	TCount = 28
	NCount = VCount * TCount
	SCount = 11172
)

var (
	LTable = []rune{'ㄱ', 'ㄲ', 'ㄴ', 'ㄷ', 'ㄸ', 'ㄹ', 'ㅁ', 'ㅂ', 'ㅃ', 'ㅅ', 'ㅆ', 'ㅇ', 'ㅈ', 'ㅉ', 'ㅊ', 'ㅋ', 'ㅌ', 'ㅍ', 'ㅎ'}
	VTable = []rune{'ㅏ', 'ㅐ', 'ㅑ', 'ㅒ', 'ㅓ', 'ㅔ', 'ㅕ', 'ㅖ', 'ㅗ', 'ㅘ', 'ㅙ', 'ㅚ', 'ㅛ', 'ㅜ', 'ㅝ', 'ㅞ', 'ㅟ', 'ㅠ', 'ㅡ', 'ㅢ', 'ㅣ'}
	TTable = []rune{0, 'ㄱ', 'ㄲ', 'ㄳ', 'ㄴ', 'ㄵ', 'ㄶ', 'ㄷ', 'ㄹ', 'ㄺ', 'ㄻ', 'ㄼ', 'ㄽ', 'ㄾ', 'ㄿ', 'ㅀ', 'ㅁ', 'ㅂ', 'ㅄ', 'ㅅ', 'ㅆ', 'ㅇ', 'ㅈ', 'ㅊ', 'ㅋ', 'ㅌ', 'ㅍ', 'ㅎ'}
)

// NormalizeBasic: 기본 정규화 (소문자 변환, 공백 제거)
func NormalizeBasic(s string) string {
	// 실제 프로덕션에서는 golang.org/x/text/unicode/norm 의 NFKC 정규화를 권장합니다.
	// 여기서는 표준 라이브러리만 사용하여 기본적인 처리를 합니다.
	s = strings.ToLower(s)
	return strings.ReplaceAll(s, " ", "")
}

// ToJamoString: 한글 문자열을 자모로 분리
func ToJamoString(str string) string {
	var out []rune
	for _, ch := range NormalizeBasic(str) {
		if ch >= SBase && ch < SBase+SCount {
			sIndex := int(ch - SBase)
			l := sIndex / NCount
			v := (sIndex % NCount) / TCount
			t := sIndex % TCount

			out = append(out, LTable[l])
			out = append(out, VTable[v])
			if t > 0 {
				out = append(out, TTable[t])
			}
		} else {
			if !unicode.IsSpace(ch) {
				out = append(out, ch)
			}
		}
	}
	return string(out)
}

// Jaro: Jaro 거리 계산
func Jaro(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	// rune slice로 변환 (한글 길이 계산을 위해 필수)
	r1, r2 := []rune(s1), []rune(s2)
	len1, len2 := len(r1), len(r2)

	if len1 == 0 || len2 == 0 {
		return 0.0
	}

	matchDist := int(math.Max(float64(len1), float64(len2))/2) - 1
	if matchDist < 0 {
		matchDist = 0
	}

	s1Matches := make([]bool, len1)
	s2Matches := make([]bool, len2)
	matches := 0.0

	for i := 0; i < len1; i++ {
		start := int(math.Max(0, float64(i-matchDist)))
		end := int(math.Min(float64(i+matchDist+1), float64(len2)))

		for j := start; j < end; j++ {
			if s2Matches[j] {
				continue
			}
			if r1[i] != r2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	k := 0
	transpositions := 0.0
	for i := 0; i < len1; i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if r1[i] != r2[k] {
			transpositions++
		}
		k++
	}
	transpositions /= 2.0

	return (matches/float64(len1) + matches/float64(len2) + (matches-transpositions)/matches) / 3.0
}

// JaroWinkler: Jaro-Winkler 유사도 계산
func JaroWinkler(s1, s2 string) float64 {
	// 자모 분리된 문자열로 Jaro 계산
	j := Jaro(s1, s2)
	if j == 0 {
		return 0.0
	}

	p := 0.1
	maxPrefix := 4.0
	
	r1, r2 := []rune(s1), []rune(s2)
	l := 0.0
	stop := math.Min(maxPrefix, math.Min(float64(len(r1)), float64(len(r2))))

	for l < stop && r1[int(l)] == r2[int(l)] {
		l++
	}

	return j + l*p*(1.0-j)
}

// Score: 최종 유사도 점수 (자모 분리 후 비교)
func Score(a, b string) float64 {
	return JaroWinkler(ToJamoString(a), ToJamoString(b))
}