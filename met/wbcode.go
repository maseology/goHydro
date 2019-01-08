package met

// WaterBalanceDataType
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
func WBcodeToMap(wdcb uint64) map[uint64]string {
	s := make(map[uint64]string)
	if wdcb&Temperature == Temperature {
		s[Temperature] = "Temperature"
	}
	if wdcb&MaxDailyT == MaxDailyT {
		s[MaxDailyT] = "MaxDailyT"
	}
	if wdcb&MinDailyT == MinDailyT {
		s[MinDailyT] = "MinDailyT"
	}
	if wdcb&Precipitation == Precipitation {
		s[Precipitation] = "Precipitation"
	}
	if wdcb&Rainfall == Rainfall {
		s[Rainfall] = "Rainfall"
	}
	if wdcb&Snowfall == Snowfall {
		s[Snowfall] = "Snowfall"
	}
	if wdcb&Snowdepth == Snowdepth {
		s[Snowdepth] = "Snowdepth"
	}
	if wdcb&SnowpackSWE == SnowpackSWE {
		s[SnowpackSWE] = "SnowpackSWE"
	}
	if wdcb&SnowMelt == SnowMelt {
		s[SnowMelt] = "SnowMelt"
	}
	if wdcb&AtmosphericYield == AtmosphericYield {
		s[AtmosphericYield] = "AtmosphericYield"
	}
	if wdcb&AtmosphericDemand == AtmosphericDemand {
		s[AtmosphericDemand] = "AtmosphericDemand"
	}
	if wdcb&Radiation == Radiation {
		s[Radiation] = "Radiation"
	}
	if wdcb&RadiationSW == RadiationSW {
		s[RadiationSW] = "RadiationSW"
	}
	if wdcb&RadiationLW == RadiationLW {
		s[RadiationLW] = "RadiationLW"
	}
	if wdcb&CloudCover == CloudCover {
		s[CloudCover] = "CloudCover"
	}
	if wdcb&RH == RH {
		s[RH] = "RH"
	}
	if wdcb&AtmosphericPressure == AtmosphericPressure {
		s[AtmosphericPressure] = "AtmosphericPressure"
	}
	if wdcb&Windspeed == Windspeed {
		s[Windspeed] = "Windspeed"
	}
	if wdcb&Windgust == Windgust {
		s[Windgust] = "Windgust"
	}
	if wdcb&WindDirection == WindDirection {
		s[WindDirection] = "WindDirection"
	}
	if wdcb&HeatDegreeDays == HeatDegreeDays {
		s[HeatDegreeDays] = "HeatDegreeDays"
	}
	if wdcb&CoolDegreeDays == CoolDegreeDays {
		s[CoolDegreeDays] = "CoolDegreeDays"
	}
	if wdcb&Unspecified23 == Unspecified23 {
		s[Unspecified23] = "Unspecified23"
	}
	if wdcb&HeadStage == HeadStage {
		s[HeadStage] = "HeadStage"
	}
	if wdcb&Flux == Flux {
		s[Flux] = "Flux"
	}
	if wdcb&UnitDischarge == UnitDischarge {
		s[UnitDischarge] = "UnitDischarge"
	}
	if wdcb&Unspecified27 == Unspecified27 {
		s[Unspecified27] = "Unspecified27"
	}
	if wdcb&Unspecified28 == Unspecified28 {
		s[Unspecified28] = "Unspecified28"
	}
	if wdcb&Unspecified29 == Unspecified29 {
		s[Unspecified29] = "Unspecified29"
	}
	if wdcb&Unspecified30 == Unspecified30 {
		s[Unspecified30] = "Unspecified30"
	}
	if wdcb&Storage == Storage {
		s[Storage] = "Storage"
	}
	if wdcb&SnowPackCover == SnowPackCover {
		s[SnowPackCover] = "SnowPackCover"
	}
	if wdcb&SnowPackLWC == SnowPackLWC {
		s[SnowPackLWC] = "SnowPackLWC"
	}
	if wdcb&SnowPackAlbedo == SnowPackAlbedo {
		s[SnowPackAlbedo] = "SnowPackAlbedo"
	}
	if wdcb&SnowSurfaceTemp == SnowSurfaceTemp {
		s[SnowSurfaceTemp] = "SnowSurfaceTemp"
	}
	if wdcb&Unspecified36 == Unspecified36 {
		s[Unspecified36] = "Unspecified36"
	}
	if wdcb&Unspecified37 == Unspecified37 {
		s[Unspecified37] = "Unspecified37"
	}
	if wdcb&DepressionWaterContent == DepressionWaterContent {
		s[DepressionWaterContent] = "DepressionWaterContent"
	}
	if wdcb&InterceptionWaterContent == InterceptionWaterContent {
		s[InterceptionWaterContent] = "InterceptionWaterContent"
	}
	if wdcb&SoilSurfaceTemp == SoilSurfaceTemp {
		s[SoilSurfaceTemp] = "SoilSurfaceTemp"
	}
	if wdcb&SoilSurfaceRH == SoilSurfaceRH {
		s[SoilSurfaceRH] = "SoilSurfaceRH"
	}
	if wdcb&SoilMoistureContent == SoilMoistureContent {
		s[SoilMoistureContent] = "SoilMoistureContent"
	}
	if wdcb&SoilMoisturePressure == SoilMoisturePressure {
		s[SoilMoisturePressure] = "SoilMoisturePressure"
	}
	if wdcb&Unspecified44 == Unspecified44 {
		s[Unspecified44] = "Unspecified44"
	}
	if wdcb&Unspecified45 == Unspecified45 {
		s[Unspecified45] = "Unspecified45"
	}
	if wdcb&Unspecified46 == Unspecified46 {
		s[Unspecified46] = "Unspecified46"
	}
	if wdcb&Unspecified47 == Unspecified47 {
		s[Unspecified47] = "Unspecified47"
	}
	if wdcb&Evaporation == Evaporation {
		s[Evaporation] = "Evaporation"
	}
	if wdcb&Transpiration == Transpiration {
		s[Transpiration] = "Transpiration"
	}
	if wdcb&Evapotranspiration == Evapotranspiration {
		s[Evapotranspiration] = "Evapotranspiration"
	}
	if wdcb&Infiltration == Infiltration {
		s[Infiltration] = "Infiltration"
	}
	if wdcb&Runoff == Runoff {
		s[Runoff] = "Runoff"
	}
	if wdcb&Recharge == Recharge {
		s[Recharge] = "Recharge"
	}
	if wdcb&TotalHead == TotalHead {
		s[TotalHead] = "TotalHead"
	}
	if wdcb&PressureHead == PressureHead {
		s[PressureHead] = "PressureHead"
	}
	if wdcb&SubSurfaceLateralFlux == SubSurfaceLateralFlux {
		s[SubSurfaceLateralFlux] = "SubSurfaceLateralFlux"
	}
	if wdcb&FluxLeft == FluxLeft {
		s[FluxLeft] = "FluxLeft"
	}
	if wdcb&FluxRight == FluxRight {
		s[FluxRight] = "FluxRight"
	}
	if wdcb&FluxFront == FluxFront {
		s[FluxFront] = "FluxFront"
	}
	if wdcb&FluxBack == FluxBack {
		s[FluxBack] = "FluxBack"
	}
	if wdcb&FluxBottom == FluxBottom {
		s[FluxBottom] = "FluxBottom"
	}
	if wdcb&FluxTop == FluxTop {
		s[FluxTop] = "FluxTop"
	}
	if wdcb&OutgoingRadiationLW == OutgoingRadiationLW {
		s[OutgoingRadiationLW] = "OutgoingRadiationLW"
	}
	if wdcb&Reserved == Reserved {
		s[Reserved] = "Reserved"
	}
	return s
}
