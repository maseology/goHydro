package richards1d

import "math"

// returns specific moisture apacity pg.120
func (pm RPM) dThetadH(h0, h1, z float64) float64 {
	psi0 := h0 + g*z
	psi1 := h1 + g*z
	if math.Abs(psi1-psi0) < 1E-5 {
		return pm.dThetadPsi(psi0)
	}
	theta0 := pm.GetTheta(psi0)
	theta1 := pm.GetTheta(psi1)
	return (theta1 - theta0) / (psi1 - psi0)
}

func (pm RPM) dThetadPsi(psi float64) float64 {
	if psi > pm.He {
		return 0.0
	}
	return -pm.GetTheta(psi) / (pm.B * psi)
}

func (pm RPM) mfpHe() float64 {
	return pm.Ks * pm.He / (-3.0/pm.B - 1.0)
}

// MFPfromTheta is needed to determine the matrix
// flux potential from water content.
func (pm RPM) MFPfromTheta(theta float64) float64 {
	return pm.mfpHe() * math.Pow(theta/pm.Ts, pm.B+3.0)
}

func (pm RPM) mfpFromPsi(psi float64) float64 {
	return pm.mfpHe() * math.Pow(psi/pm.He, -3.0/pm.B-1.0)
}

func (pm RPM) thetaFromMFP(MFP float64) float64 {
	mfphe := pm.mfpHe()
	if MFP > mfphe {
		return pm.Ts
	}
	return pm.Ts * math.Pow(MFP/mfphe, 1.0/(pm.B+3.0))
}

func meanK(k1, k2 float64) float64 {
	if k1 != k2 {
		return (k1 - k2) / math.Log(k1/k2) // logarithmic mean
	}
	return k1
}

func (pm RPM) hydraulicConductivityFromMFP(MFP float64) float64 {
	return pm.Ks * math.Pow(MFP/pm.mfpHe(), (2.0*pm.B+3.0)/(pm.B+3.0))
}
