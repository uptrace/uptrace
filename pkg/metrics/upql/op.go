package upql

type binaryOpFunc func(v1, v2 float64) float64

func addOp(v1, v2 float64) float64 {
	return v1 + v2
}

func subtractOp(v1, v2 float64) float64 {
	return v1 - v2
}

func multiplyOp(v1, v2 float64) float64 {
	return v1 * v2
}

func divideOp(v1, v2 float64) float64 {
	return v1 / v2
}

func remOp(v1, v2 float64) float64 {
	return float64(int64(v1) % int64(v2))
}

func equalOp(v1, v2 float64) float64 {
	if v1 == v2 {
		return 1
	}
	return 0
}

func gtOp(v1, v2 float64) float64 {
	if v1 > v2 {
		return 1
	}
	return 0
}

func notEqualOp(v1, v2 float64) float64 {
	if v1 != v2 {
		return 1
	}
	return 0
}

func gteOp(v1, v2 float64) float64 {
	if v1 >= v2 {
		return 1
	}
	return 0
}

func ltOp(v1, v2 float64) float64 {
	if v1 < v2 {
		return 1
	}
	return 0
}

func lteOp(v1, v2 float64) float64 {
	if v1 < v2 {
		return 1
	}
	return 0
}

func andOp(v1, v2 float64) float64 {
	if v1 != 0 && v2 != 0 {
		return v2
	}
	return 0
}

func orOp(v1, v2 float64) float64 {
	if v1 != 0 || v2 != 0 {
		return v1
	}
	return 0
}

func delta(ts *Timeseries) {
	if len(ts.Value) == 0 {
		return
	}

	prevNum := ts.Value[0]
	ts.Value[0] = 0
	value := ts.Value[1:]

	for i, num := range value {
		if delta := num - prevNum; delta >= 0 {
			value[i] = delta
		} else {
			value[i] = 0
		}
		prevNum = num
	}
}
