# goHydro

A suite of hydrological tools needed for numerical modelling

## Includes:

* **`energybal`** - A general enrgy balance scheme. Mostly used for snopack modelling.
* **`glue`** - The Generalized Likelihood Uncertainty Esimator
* **`gmet`** - a struct used to handle gridded temporal data efficiently (legacy)
* **`grid`** - a struct used to manipulate gridded data
* **`gwru`** - a Ground Water Response Unit (for hydroligcal modelling)
* **`hru`** - a Hydrologic Response Unit (for hydroligcal modelling)
* **`infiltration`** - a suite of infiltration schemes
* **`lia`** - a 2D local inertial approximation to the shallow water equation, after LISFLOOD.
* **`met`** - tools to handle timeseries data (legacy)
* **`pet`** - a suite of PET estimators
* **`porousmedia** - a set of sturcts used to define the properties of some prous medium. Includes:
  * Richards1D: a fully-implicit solution to the one-dimensional, unsteady Richards equation of variably saturated flow in porous media, based on the solution of *Bittelli, M., Campbell, G.S., and Tomei, F., (2015). Soil Physics with Python. Oxford University Press.*
* **`profile`** - essentially an ordered set of `porousmedia` making up a linear/vertical profile
* **`routing`** - a suite of topological tools optimized as a recursive set of structs
* **`snowpack`** - a snowpack modelling scheme
* **`solirrad`** - solar irradiation computer
* **`swat`** - an example of the above tools in effectively re-creating SWAT (legacy)
* **`tem`** - a Topological Elevation Model: a DEM with drain path connectivity.
* **`transfunc`** - a suite of generalized "transfer functions" used in hydrological modelling
* **`waterbudget`** - a set of classic large scale waterbudget models (for fun)
* **`wgen`** - a simple weather generator


  

