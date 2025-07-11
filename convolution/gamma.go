package convolution

import (
	"math"
)

const MAX_CONVOL_STORES = 50.

func NewGammaConvolution(alpha, beta, tsec float64) *Convolution {

	tstep := tsec / 86400.
	tmax := 4.5 * math.Pow(alpha, .6) / beta // from Raven: /*empirical estimate of tail, done in-house @UW */
	// fmt.Println(tmax)

	tmax = min(MAX_CONVOL_STORES*tstep, tmax)
	nstor := int(math.Ceil(tmax / tstep))
	if nstor == 0 {
		nstor = 1
	}
	wsum := 0.
	w := make([]float64, nstor)
	for i := range nstor {
		w[i] = gammaCumulatativeDistribution((float64(i)+1)*tstep, alpha, beta) - wsum
		wsum += w[i]
	}
	for i := range nstor {
		w[i] /= wsum
	}
	return &Convolution{w, make([]float64, nstor)}

	// max_time=min(MAX_CONVOL_STORES*tstep,max_time); //must truncate
	// N =(int)(ceil(max_time/tstep));
	// if (N==0){N=1;}
	// NN=N;
	// for (int n=0;n<N;n++)
	// {
	//   aUnitHydro[n]=LocalCumulDist((n+1)*tstep,pHRU)-sum;
	//   sum+=aUnitHydro[n];

	//   aInterval[n]=1;
	// }
	// for (int n=0;n<N;n++){aUnitHydro[n]/=sum;}
}

func gammaCumulatativeDistribution(t, alpha, beta float64) float64 {
	return incompleteGamma(beta*t, alpha) / gamma2(alpha)
}

func incompleteGamma(x, a float64) float64 {
	// modified from Raven 4.0 source code
	if x <= 0. {
		return 0.
	}
	if x > 50. {
		return incompleteGamma(49.9, a)
	}
	num, sum, prod := 1., 0., 1.
	for n := 0.; n < 100; n++ {
		if n > 0 {
			num *= x
		}
		prod *= a + n
		sum += num / prod
	}
	return sum * math.Pow(x, a) * math.Exp(-x)

	//   //cumulative distribution
	//   /// \ref from http://algolist.manual.ru/maths/count_fast/gamma_function.php
	//   const int N=100;
	//   if (x<=0  ){return 0.0;}
	//   if (x> 50 ){return IncompleteGamma(49.9,a);}
	//   double num=1.0;
	//   double sum=0.0;
	//   double prod=1.0;
	//   for (int n=0;n<N;n++){
	//     if (n>0){num*=x;}
	//     prod*=(a+n);
	//     sum+=num/prod;
	//   }
	//   return sum*pow(x,a)*exp(-x);
}

func gamma2(x float64) float64 {
	// modified from Raven 4.0 source code
	var ga float64
	g := []float64{
		1.0, 0.5772156649015329, -0.6558780715202538,
		-0.420026350340952e-1, 0.1665386113822915, -0.421977345555443e-1,
		-0.9621971527877e-2, 0.7218943246663e-2, -0.11651675918591e-2,
		-0.2152416741149e-3, 0.1280502823882e-3, -0.201348547807e-4,
		-0.12504934821e-5, 0.1133027232e-5, -0.2056338417e-6,
		0.6116095e-8, 0.50020075e-8, -0.11812746e-8,
		0.1043427e-9, 0.77823e-11, -0.36968e-11,
		0.51e-12, -0.206e-13, -0.54e-14,
		0.14e-14}

	if x > 171. {
		return 0.
	} // This value is an overflow flag.
	if x == math.Floor(x) {
		if x > 0. {
			ga = 1. // use factorial
			for i := 2.; i < x; i++ {
				ga *= i
			}
		} else {
			ga = 0.
			panic("Gamma:negative integer values not allowed")
		}
	} else {
		z := x
		r := 1.
		if math.Abs(x) > 1. {
			z = math.Abs(x)
			m := math.Floor(z)
			r = 1.
			for k := 1.; k <= m; k++ {
				r *= z - k
			}
			z -= m
		}
		gr := g[24]
		for k := 23; k >= 0; k-- {
			gr = gr*z + g[k]
		}
		ga = 1.0 / (gr * z)
		if math.Abs(x) > 1. {
			ga *= r
			if x < 0. {
				ga = -math.Pi / (x * ga * math.Sin(math.Pi*x))
			}
		}
	}
	return ga
}
