# goHydro profile

a Go struct used to define the properties of some variable porous media profile including a fully-implicit solution to the one-dimensional, unsteady Richards equation of variably saturated liquid and vapour flow in porous media, based on the solution of *Bittelli, M., Campbell, G.S., and Tomei, F., (2015). Soil Physics with Python. Oxford University Press.*

## 1D multiphase flow through porous media

Following the methodology and code examples in Bittelli etal (2015), the following presents the $\psi$-based form of the Richards (1931) equation:

$$
	% \rho_lC(\psi_m)\frac{\partial\psi_m}{\partial t}=\nabla\left[K(\psi_m)\left(\nabla\psi_m+g\mathbf{\hat{z}}\right)\right] %pg.166
	C(\psi_m)\frac{\partial\psi_m}{\partial t}=\frac{\partial}{\partial z}\left[K(\psi_m)\left(\frac{\partial\psi_m}{\partial z}+1\right)\right]
$$

where $\psi_m$ is the pressure head $[L]$, $K$ is the hydraulic conductivity $[L/T]$ and the specific moisture capacity $C(\psi_m)$ is given by

$$
	C(\psi_m)=\frac{d\theta}{d\psi_m}
$$ 
<!-- pg.120 -->

Soil water retention curve functions used follow that of Campbell (1974), *as noted by Bittelli etal (2015), the advantage of Campbell's (1974) formulation is that it is analytically integrable.*: <!-- %see also pg.172 -->

<!-- pg.104 (i believe there's a typo, or a change of sign, here taking all potentials are negative) -->
$$ 
	\theta=
	\begin{cases}
	\theta_s\left(\frac{\psi_m}{\psi_e}\right)^{-1/b} & \quad \text{if } \psi_m<\psi_e \\
	\theta_s & \quad \text{otherwise}
	\end{cases} 
$$


where $\theta$ is the volumetric water content $[L^3/L^3]$, $\theta_s$ is the saturated volumetric water content, $\psi_e$ is the air entry potential $[L]$ and $b$ is a shape parameter. Unsaturated hydraulic conductivity is given by:

$$
	K=
	\begin{cases}
	K_s\left(\frac{\psi_e}{\psi_m}\right)^{2+3/b} & \quad \text{if } \psi_m<\psi_e \\
	K_s & \quad \text{otherwise}
	\end{cases} %pg.137
$$

or in terms of volumetric water content:

$$
	K=
	\begin{cases}
	K_s\left(\frac{\theta}{\theta_s}\right)^{2b+3} & \quad \text{if } \theta<\theta_s \\
	K_s & \quad \text{otherwise}
	\end{cases}
$$

Initial conditions for the model is a degree of saturation assumed homogenous throughout the depth of the soil profile. From these above relationships, volumetric water content, matric potential, hydraulic conductivity and total head is determined. These equation allows for a solution to specific moisture capacity as:

$$
	C(\psi_m)=\frac{d\theta}{d\psi_m}=\frac{-\theta}{b\psi_m}
$$ 
<!-- pg.121 -->

Boundary conditions at the top an bottom of the soil profile can either be a specified head or a specified flux. The top can be assumed saturated, in which case the air entry potential is specified; otherwise a flux term can be set in cases where flux is less than the infiltrability of the soil. The bottom can be set to the depth of the water table, where the potential is specified as zero, or can be given a unit gradient such that it is free to drain under gravity.

more details found here: [theory.pdf](theory.pdf)

# Example

```
package main

import (
	"math"

	"github.com/maseology/goHydro/porousmedia"
	"github.com/maseology/goHydro/profile"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func main() {

	initialSe := 0.3        // initial degree of saturation
	simulationLength := 24. // hours
	
	dpth := []float64{0.2, 0.7, 1.} // profile depths
	pm := []porousmedia.PorousMedium{
		{ // coarser (less clay) textured A horizon
			Ts: 0.44,
			Tr: 0.01,
			Ks: 0.000167,
			He: -2.08,
			B:  4.74,
		},
		{ // clay loam B horizon
			Ts: 0.45,
			Tr: 0.01,
			Ks: 0.000101,
			He: -5.15,
			B:  2.59,
		},
		{ // C horizon should be parent material; here sand is assumed
			Ts: 0.37,
			Tr: 0.01,
			Ks: 0.002814,
			He: -0.73,
			B:  1.69,
		},
	}

	var p profile.Profile
	p.New(pm, dpth)
	var ps profile.State
	ps.Initialize(p, initialSe, true)

	t, f, ok := ps.Solve(simulationLength)
	linePoints("flux.png", t, f)
	if ok {
		w, z := ps.WaterContentProfile()
		linePoints("wcp.png", w, z)
	}

}

func linePoints(fp string, x, y []float64) {
	p := plot.New()

	err := plotutil.AddLinePoints(p, "v1", points(x, y))
	if err != nil {
		panic(err)
	}

	// Save the plot to a PNG file.
	if err := p.Save(12*vg.Inch, 4*vg.Inch, fp); err != nil {
		panic(err)
	}
}

func points(x, y []float64) plotter.XYs {
	if len(x) != len(y) {
		panic("error: unequal points array sizes")
	}
	pts := make(plotter.XYs, len(x))
	for i := range pts {
		if !math.IsNaN(y[i]) {
			pts[i].X = x[i]
			pts[i].Y = y[i]
		} else {
			pts[i].X = x[i]
			pts[i].Y = 0.
		}
	}
	return pts
}

```
