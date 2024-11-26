# rainrun

this go package contains a variety of lumped-parameter, catchment-based continuous rainfall-runoff models using a common interface. This package include the ability for Monte Carlo integration and optimization. 

## Included models

* The Atkinson simple storage model (Atkinson et.al., 2002, 2003; Wittenberg and Sivapalan, 1999)
* The Dawdy and O'Donnell (1965) model
* The GR4J model (Perrin et.al., 2003)
* The HBV model (Bergström, 1976, 1992; Seibert and McDonnell, 2010)
* The Manabe (1969) simple storage model with an added linear groundwater reservoir
* The multilayer capacitance model (Struthers et.al., 2003)
* The Quinn simple storage model (Quinn and Beven, 1993)
* The SIXPAR/TWOPAR model (Gupta and Sorooshian, 1983; Duan et.al., 1992)
* The simple parallel linear reservoir model (Buytaert and Beven, 2011)
* The Tank Model (Sugawara, 1995)

## References

Atkinson S.E., R.A. Woods, M. Sivapalan, 2002. Climate and landscape controls on water balance model complexity over changing timescales. Water Resource Research 38(12): 1314.

Atkinson, S.E., M. Sivapalan, N.R. Viney, R.A. Woods, 2003. Predicting space-time variability of hourly streamflow and the role of climate seasonality: Mahurangi Catchment, New Zealand. Hydrological Processes 17: 2171-2193.

Bergström, S., 1976. Development and application of a conceptual runoff model for Scandinavian catchments. SMHI RHO 7. Norrköping. 134 pp.

Bergström, S., 1992. The HBV model - its structure and applications. SMHI RH No 4. Norrköping. 35 pp.

Buytaert, W., and K. Beven, 2011. Models as multiple working hypotheses: hydrological simulation of tropical alpine . Hydrological Processes 25. pp. 1784–1799.

Dawdy, D.R., and T. O'Donnell, 1965. Mathematical Models of Catchment Behavior. Journal of Hydraulics Division, ASCE, Vol. 91, No. HY4: 123-137.

Duan, Q., S. Sorooshian, V. Gupta, 1992. Effective and Efficient Global Optimization for Conceptual Rainfall-Runoff Models. Water Resources Research 28(4): 1015-1031.

Gupta V.K., S. Sorooshian, 1983. Uniqueness and Observability of Conceptual Rainfall-Runoff Model Parameters: The Percolation Process Examined. Water Resources Research 19(1): 269-276.

Manabe, S., 1969. Climate and the Ocean Circulation 1: The Atmospheric Circulation and The Hydrology of the Earth's Surface. Monthly Weather Review 97(11): 739-744.

Perrin C., C. Michel, V. Andreassian, 2003. Improvement of a parsimonious model for streamflow simulation. Journal of Hydrology 279: 275-289.

Quinn P.F., K.J. Beven, 1993. Spatial and temporal predictions of soil moisture dynamics, runoff, variable source areas and evapotranspiration for Plynlimon, mid-Wales. Hydrological Processes 7: 425-448.

Seibert, J. and J.J. McDonnell, 2010. Land-cover impacts on streamflow: a change-detection modelling approach that incorporates parameter uncertainty. Hydrological Sciences Journal 55(3): 316-332.

Struthers, I., C. Hinz, M. Sivapalan, G. Deutschmann, F. Beese, R. Meissner, 2003. Modelling the water balance of a free-draining lysimeter using the downward approach. Hydrological Processes (17): 2151-2169.

Sugawara, M. (1995). Tank model. In: V.P. Singh (Ed.), Computer models of watershed hydrology. Water Resources Publications, Highlands Ranch, Colorado.

Wittenberg H., M. Sivapalan, 1999. Watershed groundwater balance equation using streamflow recession analysis and baseflow separation. Journal of Hydrology 219: 20-33.