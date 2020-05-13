package met

// WaterBalanceDataType (bit-wise type code)
const (
	Temperature = 1 << iota
	MaxDailyT
	MinDailyT
	Precipitation
	Rainfall
	Snowfall
	Snowdepth
	SnowpackSWE
	SnowMelt
	AtmosphericYield
	AtmosphericDemand
	Radiation
	RadiationSW
	RadiationLW
	CloudCover
	RH
	AtmosphericPressure
	Windspeed
	Windgust
	WindDirection
	HeatDegreeDays
	CoolDegreeDays
	Unspecified23
	HeadStage
	Flux
	UnitDischarge
	Unspecified27
	Unspecified28
	Unspecified29
	Unspecified30
	Storage
	SnowPackCover
	SnowPackLWC
	SnowPackAlbedo
	SnowSurfaceTemp
	Unspecified36
	Unspecified37
	DepressionWaterContent
	InterceptionWaterContent
	SoilSurfaceTemp
	SoilSurfaceRH
	SoilMoistureContent
	SoilMoisturePressure
	Unspecified44
	Unspecified45
	Unspecified46
	Unspecified47
	Evaporation
	Transpiration
	Evapotranspiration
	Infiltration
	Runoff
	Recharge
	TotalHead
	PressureHead
	SubSurfaceLateralFlux
	FluxLeft
	FluxRight
	FluxFront
	FluxBack
	FluxBottom
	FluxTop
	OutgoingRadiationLW
	Reserved
)

// WBcodeToMap converts a wbdc into a list of metrics
func WBcodeToMap(wbdc uint64) map[uint64]string {
	s := make(map[uint64]string)
	if wbdc&Temperature == Temperature {
		s[Temperature] = "Temperature"
	}
	if wbdc&MaxDailyT == MaxDailyT {
		s[MaxDailyT] = "MaxDailyT"
	}
	if wbdc&MinDailyT == MinDailyT {
		s[MinDailyT] = "MinDailyT"
	}
	if wbdc&Precipitation == Precipitation {
		s[Precipitation] = "Precipitation"
	}
	if wbdc&Rainfall == Rainfall {
		s[Rainfall] = "Rainfall"
	}
	if wbdc&Snowfall == Snowfall {
		s[Snowfall] = "Snowfall"
	}
	if wbdc&Snowdepth == Snowdepth {
		s[Snowdepth] = "Snowdepth"
	}
	if wbdc&SnowpackSWE == SnowpackSWE {
		s[SnowpackSWE] = "SnowpackSWE"
	}
	if wbdc&SnowMelt == SnowMelt {
		s[SnowMelt] = "SnowMelt"
	}
	if wbdc&AtmosphericYield == AtmosphericYield {
		s[AtmosphericYield] = "AtmosphericYield"
	}
	if wbdc&AtmosphericDemand == AtmosphericDemand {
		s[AtmosphericDemand] = "AtmosphericDemand"
	}
	if wbdc&Radiation == Radiation {
		s[Radiation] = "Radiation"
	}
	if wbdc&RadiationSW == RadiationSW {
		s[RadiationSW] = "RadiationSW"
	}
	if wbdc&RadiationLW == RadiationLW {
		s[RadiationLW] = "RadiationLW"
	}
	if wbdc&CloudCover == CloudCover {
		s[CloudCover] = "CloudCover"
	}
	if wbdc&RH == RH {
		s[RH] = "RH"
	}
	if wbdc&AtmosphericPressure == AtmosphericPressure {
		s[AtmosphericPressure] = "AtmosphericPressure"
	}
	if wbdc&Windspeed == Windspeed {
		s[Windspeed] = "Windspeed"
	}
	if wbdc&Windgust == Windgust {
		s[Windgust] = "Windgust"
	}
	if wbdc&WindDirection == WindDirection {
		s[WindDirection] = "WindDirection"
	}
	if wbdc&HeatDegreeDays == HeatDegreeDays {
		s[HeatDegreeDays] = "HeatDegreeDays"
	}
	if wbdc&CoolDegreeDays == CoolDegreeDays {
		s[CoolDegreeDays] = "CoolDegreeDays"
	}
	if wbdc&Unspecified23 == Unspecified23 {
		s[Unspecified23] = "Unspecified23"
	}
	if wbdc&HeadStage == HeadStage {
		s[HeadStage] = "HeadStage"
	}
	if wbdc&Flux == Flux {
		s[Flux] = "Flux"
	}
	if wbdc&UnitDischarge == UnitDischarge {
		s[UnitDischarge] = "UnitDischarge"
	}
	if wbdc&Unspecified27 == Unspecified27 {
		s[Unspecified27] = "Unspecified27"
	}
	if wbdc&Unspecified28 == Unspecified28 {
		s[Unspecified28] = "Unspecified28"
	}
	if wbdc&Unspecified29 == Unspecified29 {
		s[Unspecified29] = "Unspecified29"
	}
	if wbdc&Unspecified30 == Unspecified30 {
		s[Unspecified30] = "Unspecified30"
	}
	if wbdc&Storage == Storage {
		s[Storage] = "Storage"
	}
	if wbdc&SnowPackCover == SnowPackCover {
		s[SnowPackCover] = "SnowPackCover"
	}
	if wbdc&SnowPackLWC == SnowPackLWC {
		s[SnowPackLWC] = "SnowPackLWC"
	}
	if wbdc&SnowPackAlbedo == SnowPackAlbedo {
		s[SnowPackAlbedo] = "SnowPackAlbedo"
	}
	if wbdc&SnowSurfaceTemp == SnowSurfaceTemp {
		s[SnowSurfaceTemp] = "SnowSurfaceTemp"
	}
	if wbdc&Unspecified36 == Unspecified36 {
		s[Unspecified36] = "Unspecified36"
	}
	if wbdc&Unspecified37 == Unspecified37 {
		s[Unspecified37] = "Unspecified37"
	}
	if wbdc&DepressionWaterContent == DepressionWaterContent {
		s[DepressionWaterContent] = "DepressionWaterContent"
	}
	if wbdc&InterceptionWaterContent == InterceptionWaterContent {
		s[InterceptionWaterContent] = "InterceptionWaterContent"
	}
	if wbdc&SoilSurfaceTemp == SoilSurfaceTemp {
		s[SoilSurfaceTemp] = "SoilSurfaceTemp"
	}
	if wbdc&SoilSurfaceRH == SoilSurfaceRH {
		s[SoilSurfaceRH] = "SoilSurfaceRH"
	}
	if wbdc&SoilMoistureContent == SoilMoistureContent {
		s[SoilMoistureContent] = "SoilMoistureContent"
	}
	if wbdc&SoilMoisturePressure == SoilMoisturePressure {
		s[SoilMoisturePressure] = "SoilMoisturePressure"
	}
	if wbdc&Unspecified44 == Unspecified44 {
		s[Unspecified44] = "Unspecified44"
	}
	if wbdc&Unspecified45 == Unspecified45 {
		s[Unspecified45] = "Unspecified45"
	}
	if wbdc&Unspecified46 == Unspecified46 {
		s[Unspecified46] = "Unspecified46"
	}
	if wbdc&Unspecified47 == Unspecified47 {
		s[Unspecified47] = "Unspecified47"
	}
	if wbdc&Evaporation == Evaporation {
		s[Evaporation] = "Evaporation"
	}
	if wbdc&Transpiration == Transpiration {
		s[Transpiration] = "Transpiration"
	}
	if wbdc&Evapotranspiration == Evapotranspiration {
		s[Evapotranspiration] = "Evapotranspiration"
	}
	if wbdc&Infiltration == Infiltration {
		s[Infiltration] = "Infiltration"
	}
	if wbdc&Runoff == Runoff {
		s[Runoff] = "Runoff"
	}
	if wbdc&Recharge == Recharge {
		s[Recharge] = "Recharge"
	}
	if wbdc&TotalHead == TotalHead {
		s[TotalHead] = "TotalHead"
	}
	if wbdc&PressureHead == PressureHead {
		s[PressureHead] = "PressureHead"
	}
	if wbdc&SubSurfaceLateralFlux == SubSurfaceLateralFlux {
		s[SubSurfaceLateralFlux] = "SubSurfaceLateralFlux"
	}
	if wbdc&FluxLeft == FluxLeft {
		s[FluxLeft] = "FluxLeft"
	}
	if wbdc&FluxRight == FluxRight {
		s[FluxRight] = "FluxRight"
	}
	if wbdc&FluxFront == FluxFront {
		s[FluxFront] = "FluxFront"
	}
	if wbdc&FluxBack == FluxBack {
		s[FluxBack] = "FluxBack"
	}
	if wbdc&FluxBottom == FluxBottom {
		s[FluxBottom] = "FluxBottom"
	}
	if wbdc&FluxTop == FluxTop {
		s[FluxTop] = "FluxTop"
	}
	if wbdc&OutgoingRadiationLW == OutgoingRadiationLW {
		s[OutgoingRadiationLW] = "OutgoingRadiationLW"
	}
	if wbdc&Reserved == Reserved {
		s[Reserved] = "Reserved"
	}
	return s
}
