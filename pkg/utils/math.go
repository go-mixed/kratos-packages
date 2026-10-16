package utils

import (
	"math"

	"golang.org/x/exp/constraints"
)

type Position struct {
	Offset int
	Length int
}

func (p Position) EndOffset() int {
	return p.Offset + p.Length
}

type Positions []Position

func (p Positions) Offset() int {
	if len(p) == 0 {
		return 0
	}
	return p[0].Offset
}

func (p Positions) EndOffset() int {
	l := len(p)
	if l == 0 {
		return 0
	}
	return p[l-1].EndOffset()
}

// Gcd 获取最大公约数（Greatest Common Divisor）
func Gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// Clamp 对值进行范围限制
func Clamp[T constraints.Integer | constraints.Float](val T, min, max T) T {
	if val < min {
		return min
	} else if val > max {
		return max
	}
	return val
}

// RoundToEven 按指定小数位进行银行家舍入。
//
// 参数 places 表示保留的小数位数，例如 places=4 表示保留 4 位小数。
// 使用 math.RoundToEven 是为了在刚好处于 0.5 的边界时舍入到最近的偶数，
// 避免大量金额连续四舍五入时总是向上偏移。
// places 小于等于 0 时按整数舍入。
func RoundToEven[T constraints.Float](value T, places int) T {
	if places <= 0 {
		return T(math.RoundToEven(float64(value)))
	}

	scale := math.Pow10(places)
	return T(math.RoundToEven(float64(value)*scale) / scale)
}

type Number interface {
	constraints.Integer | constraints.Float
}

// NumberEqual 判断两个数字是否相等，考虑浮点数精度问题。
func NumberEqual[A, B Number](a A, b B, epsilon float64) bool {
	af := float64(a)
	bf := float64(b)
	return af == bf || math.Abs(af-bf) <= epsilon
}
