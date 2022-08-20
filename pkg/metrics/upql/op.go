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
