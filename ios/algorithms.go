package ios

import (
	"fmt"
)

const ASIZE = 256

func preBmBc(x []byte, m int, bmBc [ASIZE]int) [ASIZE]int {
	for i := 0; i < ASIZE; i++ {
		bmBc[i] = m
	}
	for i := 0; i < m -1; i++ {
		bmBc[x[i]] = m-i-1
	}
	return bmBc
}

func suffixes(x []byte, m int, suff []int) []int {
	var f, g int
	suff[m-1] = m
	g = m - 1
	for i := m - 2; i >= 0; i-- {
		if i > g && suff[i+m-1-f] < i - g {
			suff[i] = suff[i+m-1-f]
		} else {
			if i < g {
				g = i
			}
			f = i
			for g >= 0 && x[g] == x[g + m - 1 -f] {
				g--
			}
			suff[i] = f - g
		}
	}
	return suff
}

func preBmGs(x []byte, m int, bmGs []int) []int {
	suff := make([]int, len(x))
	suff = suffixes(x, m, suff)
	for i := 0; i < m; i++ {
		bmGs[i] = m
	}
	j := 0
	for i := m-1; i >= 0; i-- {
		if suff[i] == i+1 {
			for ; j < m - 1 -i; j++ {
				if bmGs[j] == m {
					bmGs[j] = m -1-i
				}
			}
		}
	}
	for i := 0; i <= m - 2; i++ {
		bmGs[m - 1 - suff[i]] = m - 1 - i
	}
	return bmGs
}

func TBM(x []byte, y []byte) int {
	var bcShift, i, j, shift, u, v, turboShift int
	m := len(x)
	n := len(y)
	var bmGs = make([]int, m)
	var bmBc [ASIZE]int
	/* Preprocessing */
	preBmGs(x, m, bmGs)
	preBmBc(x, m, bmBc)
	/* Searching */
	shift = m
	for j <= n-m {
		i = m - 1
		for i >= 0 && x[i+j] == y[i] {
			i--
			if u != 0 && i == m-1-shift {
				i -= u
			}
		}
		if i < 0 {
			shift = bmGs[0]
			u = m - shift
			return j
		} else {
			v = m - 1 - i
			turboShift = u - v
			bcShift = bmBc[y[i]] - m + 1 + i
			shift = max(turboShift, bcShift)
			shift = max(shift, bmGs[i])
			if shift == bmGs[i] {
				u = min(m-shift, v)
			} else {
				if turboShift < bcShift {
					shift = max(shift, u+1)
				}
				u = 0
			}
		}
		j += shift
	}
	if j == n {
		return -1
	}
	return j - 1
}


func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	data := []byte("foobardickcissel-Libellula_dickcissel-Libellula_unicum-cyfoobarentoblast_loopful-baler_prolusory-finitive_Cereus-premorse_poolroot-krocket_staree-choristoblastoma_Osmanie-incidently_gignitive-Amblyomma_perishment-knavery_gignitive-Amblyomma_torpid-pigeonable_dickcissel-Libellula_dickcissel-Libellula_unicum-cytoblast_loopful-baler_prolusory-finitive_Cereus-premorse_poolroot-krocket_staree-cfoobarhoristoblastoma_Osmanie-incidently_gignitive-Amblyomma_perishment-knavery_gignitive-Amfoobarblyomma_torpid-pigeonable_dickcissel-Libellula_dickcissel-Libellula_unicum-cytoblast_loopful-baler_prolusory-finitive_Cereus-premorse_poolroot-krocket_staree-chorisfoobartoblastoma_Osmanie-incidently_gignitive-Amblyomma_perishment-knavery_gifoobargnitive-Amblyomma_torpid-pigeonable_")
	token := []byte("incidently_gignitive")
	fmt.Println(TBM(data, token))
}
