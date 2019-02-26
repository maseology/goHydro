package lia

import (
"github.com/maseology/goHydro/grid"
)

func (m *LIA) build(gd grid.Definition, z ,h,n map[int]float64) {
	If Not GD.IsUniform Then Stop
	_dx = GD.CellWidth(0)
	_gf = New Grid.Face(GD)
	_f = New Dictionary(Of Integer, _face)
	_fxr = New Dictionary(Of Integer, Integer())
	_bf = New Dictionary(Of Integer, Boolean)
	_n = New Dictionary(Of Integer, _node)
	_s = New Dictionary(Of Integer, _state)
	For Each n In Mannings_n
		_n.Add(n.Key, New _node With {.Elevation = z(n.Key), .Head = h0(n.Key), .Mannings_n = n.Value, .FaceID = _gf.CellFace(n.Key)})
	Next
	For i = 0 To _gf.nFaces - 1
		Dim fc1 As New _face(_gf, i)
		If fc1.IsInactive Then Continue For
		_bf.Add(i, fc1.IsBoundary)
		_f.Add(i, fc1)
		If Not fc1.IsBoundary Then
			_fxr.Add(i, fc1.IdColl)
			_s.Add(i, New _state)
		End If
	Next
	For Each f In _f
		If _bf(f.Key) Then Continue For
		Dim n = f.Value.NodeIDs
		f.Value.Initialize(_n(n(0)), _n(n(1)), _theta, _dx)
	Next
}

// NewTest constructor
func  NewTest() LIA {
	var m LIA
	m.alpha = 0.7
	m.theta = 0.7
	m.tresid = 0.00001
	return m
}

// NewDEM constructs LIA from a DEM
func  NewDEM(g grid.Real) LIA {
	Dim dicz As New Dictionary(Of Integer, Double), dicn As New Dictionary(Of Integer, Double), dich As New Dictionary(Of Integer, Double)
	For cid = 0 To DEM.GridDefinition.NumCells - 1
		dicz.Add(cid, DEM.Value(-9999, cid))
		dicn.Add(cid, 0.05)
		dich.Add(cid, dicz(cid))
	Next
	Me.Build(DEM.GridDefinition, dicz, dich, dicn)
}

// NewDEMn constructs LIA from a DEM and constant mannings n
func NewDEMn(g grid.Real, n float64) LIA {
	Dim dicz As New Dictionary(Of Integer, Double), dicn As New Dictionary(Of Integer, Double), dich As New Dictionary(Of Integer, Double)
	For Each cid In DEM.GridDefinition.Actives(True)
		dicz.Add(cid, DEM.Value(-9999, cid))
		dicn.Add(cid, Mannings_n)
		//dich.Add(cid, 0.00001 + dicz(cid))
		dich.Add(cid, dicz(cid))
	Next
	Me.Build(DEM.GridDefinition, dicz, dich, dicn)
}

// NewDEMns constructs LIA from a DEM and mannings n field
func NewDEMns(g , n grid.Real) LIA {
	Dim dicz As New Dictionary(Of Integer, Double), dich As New Dictionary(Of Integer, Double)
	For cid = 0 To DEM.GridDefinition.NumCells - 1
		dicz.Add(cid, DEM.Value(-9999, cid))
		dich.Add(cid, dicz(cid))
	Next
	Me.Build(DEM.GridDefinition, dicz, dich, Mannings_n)
}

// NewDEMhns constructs LIA from a DEM and mannings n field, with initial heads
func NewDEMns(g ,h, n grid.Real) LIA {
	Dim dicz As New Dictionary(Of Integer, Double)
	For cid = 0 To DEM.GridDefinition.NumCells - 1
		dicz.Add(cid, DEM.Value(-9999, cid))
	Next
	Me.Build(DEM.GridDefinition, dicz, h0, Mannings_n)
}