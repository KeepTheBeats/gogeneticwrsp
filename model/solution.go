package model

// Solution for scheduling applications to clouds
type Solution struct {
	SchedulingResult []int `json:"schedulingResult"`
}

// SolutionCopy deep copy a solution
func SolutionCopy(src Solution) Solution {
	var dst Solution = src
	dst.SchedulingResult = make([]int, len(src.SchedulingResult))
	copy(dst.SchedulingResult, src.SchedulingResult)
	return dst
}
