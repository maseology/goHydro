# goHydro

A suite of hydrological tools and [structs](https://go.dev/tour/moretypes/2) used with numerical modelling

## Includes:

* **`channel`** -- a set of geometries used to compute depth-area relationships.
    * trapezoid
    * *more to come..*
* **`convolution`** -- a set of convolution models/transfer functions used to simulate attenuation:
    * Snyder
    * Triangular
* **`energybal`** -- a general energy balance scheme. Mostly used for snowpack modelling.
* **`glue`** -- a *Generalized Likelihood Uncertainty Estimator* struct that is sorting-safe.
* **`grid`** -- a set of Go struct used to manipulate gridded data.
* **`gwru`** -- a Ground Water Response Unit (for hydrological modelling)--mainly a distributed application of TOPMODEL.
* **`hechms`** -- the [HEC-HMS model](https://www.hec.usace.army.mil/software/hec-hms/) (partially) rebuilt in Go.
* **`hru`** -- a Hydrologic Response Unit struct.
* **`hyetograph`** -- a set of synthetic hyetographs used in hydrology:
    * SCSII
    * Timmins design storm
    * Unit/uniform
* **`infiltration`** -- a suite of infiltration schemes:
    * Curve Number method
    * *more to come..*
* **`mesh`** -- a set of Go struct used to manipulate unstructured data (e.g., TINs).
* **`pet`** -- a suite of potential evapotranspiration estimators:
    * Makkink
    * Penman
    * Penman-Monteith
    * simple Sine curve
    * *more to come..*
* **`porousmedia`** -- A struct used to hold common subsurface material properties.
* **`profile`** -- an ordered set of `porousmedia` making up a linear/vertical subsurface profile.
* **`rainrun`** -- a variety of lumped-parameter, catchment-based continuous rainfall-runoff models using a common interface. This package include the ability for Monte Carlo integration and optimization. Models include:
    * Atkinson (2003)
    * Dawdy and O'Donnel (1965)
    * GR4J (2003)
    * HBV (1976)
    * Manabe (1969)
    * Multi-Layer capacitance (2003)
    * Quinn (1993)
    * Sixpar (1983)
    * SPLR (2011)
* **`routing`** -- a suite of topological tools optimized as a recursive set of Go structs.
* **`snowpack`** -- a snowpack modelling scheme:
    * Cold-content factor
    * Degree-day factor
    * Energy balance *(to come)*
* **`solirrad`** -- solar irradiation estimation.
* **`swat`** -- [the SWAT model](https://swat.tamu.edu/) effectively re-created using goHydro.
* **`tem`** -- a Topological Elevation Model: a DEM with drain path connectivity.
* **`waterbudget`** -- a set of classic large scale water-budget models:
    * ABC model
    * ABCD model
    * Budyko model
    * NOPEX6
    * Palmer
    * Thornthwaite and Mather
* **`wgen`** -- a simple weather generator.

### Legacy (no future updates)/Deprecated

* **`gmet`** -- a struct used to handle gridded temporal climate data efficiently (legacy).
* **`met`** -- tools to handle time series data (legacy).
* **`lia`** -- an explicit, 2D local inertial approximation to the shallow water equation, following LISFLOOD-FP. 
    * *deprecated: moved to [__goLIA__](https://github.com/maseology/goLIA).*
* **`porousmedia`** -- a Go struct used to define the properties of some variable porous media profile including a fully-implicit solution to the one-dimensional, unsteady Richards equation of variably saturated liquid and vapour flow in porous media, based on the solution of *Bittelli, M., Campbell, G.S., and Tomei, F., (2015). Soil Physics with Python. Oxford University Press.*
    * *deprecated: moved to [__goVSF__](https://github.com/maseology/goVSF).*