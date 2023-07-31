package encoding

import "math"

func DouglasPeucker(points []Point, epsilon int) []Point {
	// 找到最大阈值点
	var maxH float64 = 0
	var index = 0
	end := len(points)
	for i := 1; i < end-1; i++ {
		h := calH(points[i], points[0], points[end-1])
		if h > maxH {
			maxH = h
			index = i
		}
	}

	var result []Point
	// 如果存在最大阈值点，就进行递归遍历出所有最大阈值点
	if maxH > float64(epsilon) {
		var leftPoints []Point
		var rightPoints []Point
		// 分别提取出左曲线和右曲线的坐标点
		for i := 0; i < end; i++ {
			if i <= index {
				leftPoints = append(leftPoints, points[i])
				if i == index {
					rightPoints = append(rightPoints, points[i])
				}
			} else {
				rightPoints = append(rightPoints, points[i])
			}
		}
		// 分别保存两边遍历的结果
		var leftResult []Point
		var rightResult []Point
		leftResult = DouglasPeucker(leftPoints, epsilon)
		rightResult = DouglasPeucker(rightPoints, epsilon)
		// 将两边的结果整合
		rightResult = rightResult[1:] //移除重复点
		leftResult = append(leftResult, rightResult[0:]...)
		result = leftResult
	} else { // 如果不存在最大阈值点则返回当前遍历的子曲线的起始点
		result = append(result, points[0])
		result = append(result, points[end-1])
	}
	return result
}

/**
 * 计算点到直线的距离
 *
 * @param p
 * @param s
 * @param e
 * @return
 */
func calH(p Point, s Point, e Point) float64 {
	AB := getDistanceByTwoPoint(s, e)
	CB := getDistanceByTwoPoint(p, s)
	CA := getDistanceByTwoPoint(p, e)
	S := helen(CB, CA, AB)
	H := 2 * S / AB
	return H
}

/**
 * 计算两点之间的距离
 *
 * @param p1
 * @param p2
 * @return
 */
func getDistanceByTwoPoint(p1 Point, p2 Point) float64 {
	x1 := p1.x
	y1 := p1.y
	x2 := p2.x
	y2 := p2.y
	xy := math.Sqrt(float64((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)))
	return xy
}

/**
 * 海伦公式，已知三边求三角形面积
 *
 * @param cB
 * @param cA
 * @param aB
 * @return 面积
 */
func helen(CB float64, CA float64, AB float64) float64 {
	p := (CB + CA + AB) / 2
	S := math.Sqrt(p * (p - CB) * (p - CA) * (p - AB))
	return S
}
