package grid

// https://homepages.inf.ed.ac.uk/rbf/HIPR2/gsmooth.htm
// center cell was modified from .15018 to .15020 such that the filter summed to 1.
var FilterGaussianSmoothing = [][]float64{
	{0.00366, 0.01465, 0.02564, 0.01465, 0.00366},
	{0.01465, 0.05861, 0.09524, 0.05861, 0.01465},
	{0.02564, 0.09524, 0.15020, 0.09524, 0.02564},
	{0.01465, 0.05861, 0.09524, 0.05861, 0.01465},
	{0.00366, 0.01465, 0.02564, 0.01465, 0.00366},
}
