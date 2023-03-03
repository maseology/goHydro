# goHydro

A suite of hydrological tools needed for numerical modelling

## Includes:

* **`energybal`** - A general energy balance scheme. Mostly used for snowpack modelling.
* **`glue`** - a *Generalized Likelihood Uncertainty Estimator* struct that is sorting-safe
* **`gmet`** - a struct used to handle gridded temporal data efficiently (legacy)
* **`grid`** - a struct used to manipulate gridded data
* **`gwru`** - a Ground Water Response Unit (for hydrological modelling)
 * mainly a distributed application of TOPMODEL
* **`hru`** - a Hydrologic Response Unit (for hydrological modelling)
* **`infiltration`** - a suite of infiltration schemes
 * Curve Number method
* **`lia`** - an explicit, 2D local inertial approximation to the shallow water equation, following LISFLOOD-FP. 
    * *deprecated: moved to [__goLIA__](https://github.com/maseology/goLIA).*
* **`met`** - tools to handle time series data (legacy)
* **`pet`** - a suite of PET estimators
* **`porousmedia`** - a Go struct used to define the properties of some variable porous media profile including a fully-implicit solution to the one-dimensional, unsteady Richards equation of variably saturated liquid and vapour flow in porous media, based on the solution of *Bittelli, M., Campbell, G.S., and Tomei, F., (2015). Soil Physics with Python. Oxford University Press.*
    * *deprecated: moved to [__goVSF__](https://github.com/maseology/goVSF).*
* **`profile`** - an ordered set of `porousmedia` making up a linear/vertical subsurface profile.
* **`rainrun`** - a variety of lumped-parameter, catchment-based continuous rainfall-runoff models using a common interface. This package include the ability for Monte Carlo integration and optimization. 
* **`routing`** - a suite of topological tools optimized as a recursive set of structs
* **`snowpack`** - a snowpack modelling scheme
* **`solirrad`** - solar irradiation computer
* **`swat`** - SWAT model effectively re-created using goHydro
* **`tem`** - a Topological Elevation Model: a DEM with drain path connectivity.
* **`transfunc`** - a suite of generalized "transfer functions" used in hydrological modelling
* **`waterbudget`** - a set of classic large scale waterbudget models (for fun)
* **`wgen`** - a simple weather generator
