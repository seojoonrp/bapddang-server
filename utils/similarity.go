// utils/similarity.go
// 검증 안해봄

package utils

import (
	"sort"
	"strings"
	"unicode"

	"github.com/xrash/smetrics"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// --- 1. Jamo 분해 로직 (JS의 toJamoString 포팅) ---
var (
	LTable = []rune{'ㄱ', 'ㄲ', 'ㄴ', 'ㄷ', 'ㄸ', 'ㄹ', 'ㅁ', 'ㅂ', 'ㅃ', 'ㅅ', 'ㅆ', 'ㅇ', 'ㅈ', 'ㅉ', 'ㅊ', 'ㅋ', 'ㅌ', 'ㅍ', 'ㅎ'}
	VTable = []rune{'ㅏ', 'ㅐ', 'ㅑ', 'ㅒ', 'ㅓ', 'ㅔ', 'ㅕ', 'ㅖ', 'ㅗ', 'ㅘ', 'ㅙ', 'ㅚ', 'ㅛ', 'ㅜ', 'ㅝ', 'ㅞ', 'ㅟ', 'ㅠ', 'ㅡ', 'ㅢ', 'ㅣ'}
	TTable = []rune{' ', 'ㄱ', 'ㄲ', 'ㄳ', 'ㄴ', 'ㄵ', 'ㄶ', 'ㄷ', 'ㄹ', 'ㄺ', 'ㄻ', 'ㄼ', 'ㄽ', 'ㄾ', 'ㄿ', 'ㅀ', 'ㅁ', 'ㅂ', 'ㅄ', 'ㅅ', 'ㅆ', 'ㅇ', 'ㅈ', 'ㅊ', 'ㅋ', 'ㅌ', 'ㅍ', 'ㅎ'}

	SBase  rune = 0xac00
	LCount rune = 19
	VCount rune = 21
	TCount rune = 28
	NCount rune = VCount * TCount
	SCount rune = 11172
)

// normalizeBasic은 JS의 normalizeBasic을 Go로 구현한 것입니다.
func normalizeBasic(s string) string {
	// NFKC 정규화
	t := transform.Chain(norm.NFKC, runes.Remove(runes.In(unicode.Mn)))
	result, _, _ := transform.String(t, s)
	// 소문자 변환 및 공백 제거
	result = strings.ToLower(result)
	result = strings.Join(strings.Fields(result), "") // JS의 replace(/\s+/g, "")와 동일
	return result
}

// ToJamoString은 JS의 toJamoString을 Go로 구현한 것입니다.
func ToJamoString(str string) string {
	var out strings.Builder
	normalized := normalizeBasic(str)

	for _, ch := range normalized {
		if ch >= SBase && ch < SBase+SCount {
			SIndex := ch - SBase
			L := SIndex / NCount
			V := (SIndex % NCount) / TCount
			T := SIndex % TCount

			out.WriteRune(LTable[L])
			out.WriteRune(VTable[V])
			if T > 0 {
				out.WriteRune(TTable[T])
			}
		} else {
			out.WriteRune(ch)
		}
	}
	return out.String()
}

// --- 2. 점수 계산 및 매칭 로직 (JS의 score, getBestMatches 포팅) ---

// Score는 Jamo 분해 후 Jaro-Winkler 점수를 계산합니다.
func Score(a, b string) float64 {
	jamoA := ToJamoString(a)
	jamoB := ToJamoString(b)
	// JaroWinkler(jamoA, jamoB, p = 0.1, maxPrefix = 4)
	return smetrics.JaroWinkler(jamoA, jamoB, 0.1, 4)
}

// MatchResult는 유사도 검사 결과를 담습니다.
type MatchResult struct {
	Name  string  `json:"name"`
	Score float64 `json:"score"`
}

// GetBestMatches는 JS의 getBestMatches를 Go로 구현한 것입니다.
func GetBestMatches(inputName string, candidates []string, k int) (best *MatchResult, top []MatchResult) {
	if strings.TrimSpace(inputName) == "" || len(candidates) == 0 {
		return nil, []MatchResult{}
	}

	matches := make([]MatchResult, len(candidates))
	for i, name := range candidates {
		matches[i] = MatchResult{
			Name:  name,
			Score: Score(inputName, name),
		}
	}

	// 점수가 높은 순으로 정렬
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > 0 {
		best = &matches[0]
	}
	if len(matches) < k {
		k = len(matches)
	}
	
	top = matches[:k]
	return best, top
}