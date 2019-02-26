package lia



Public Function Solve() As Dictionary(Of Integer, Double)
' steady-state
Dim sf As New Dictionary(Of Integer, _face)
For Each f In _f
	If _bf(f.Key) Then Continue For
	sf.Add(f.Key, f.Value)
Next
100:        Me.SetCurrentState()
_tacum += _dt
'For Each f In sf
'    f.Value.UpdateFlux(_s(f.Key), _dt)
'Next
Parallel.ForEach(sf, Sub(f) f.Value.UpdateFlux(_s(f.Key), _dt))
Dim r = Me.UpdateHeads
Console.WriteLine("{0:0.00000}  {1:0.0000}", _tacum, r)
If Math.Abs(r) > _tresid Then GoTo 100
Dim dicOut As New Dictionary(Of Integer, Double)
For Each n In _n
	If n.Key >= _gf.nCells Then Exit For ' ghost node boundary condition
	dicOut.Add(n.Key, n.Value.Head)
Next
Return dicOut
End Function
Public Function Solve(TimeStepSec As Double) As Dictionary(Of Integer, Double)
_tacum = 0.0
_dt = TimeStepSec
Dim sf As New Dictionary(Of Integer, _face)
For Each f In _f
	If _bf(f.Key) Then Continue For
	sf.Add(f.Key, f.Value)
Next
Do
	Me.SetCurrentState()
	_tacum += _dt
	If _tacum > TimeStepSec Then
		_dt -= _tacum - TimeStepSec
		_tacum = TimeStepSec
	End If
	'For Each f In sf
	'    f.Value.UpdateFlux(_s(f.Key), _dt)
	'Next
	Parallel.ForEach(sf, Sub(f) f.Value.UpdateFlux(_s(f.Key), _dt))
	Me.pUpdateHeads()
	Console.Write(".")
Loop Until _tacum = TimeStepSec

Dim dicOut As New Dictionary(Of Integer, Double)
For Each n In _n
	If n.Key >= _gf.nCells Then Exit For ' ghost node boundary condition
	dicOut.Add(n.Key, n.Value.Head)
Next
Return dicOut
End Function
Public Function Velocities() As Dictionary(Of Integer, Double)
Dim dicOut As New Dictionary(Of Integer, Double)
For Each n In _n
	If n.Key >= _gf.nCells Then Exit For ' ghost node boundary condition
	With n.Value
		If .Depth > 0.002 * _dx Then dicOut.Add(n.Key, Math.Sqrt(((_f(.FaceID(2)).Flux + _f(.FaceID(0)).Flux) / 2.0) ^ 2.0 + ((_f(.FaceID(3)).Flux + _f(.FaceID(1)).Flux) / 2.0) ^ 2.0) / .Depth) Else dicOut.Add(n.Key, 0.0)
	End With
Next
Return dicOut
End Function

Private Sub SetCurrentState()
Dim dmax = Double.MinValue
For Each n In _n.Values
	If n.Depth > dmax Then dmax = n.Depth
Next
If dmax > 0.0 Then _dt = _alpha * _dx / Math.Sqrt(9.80665 * dmax) ' eq.12
'For Each s In _s
'    With s.Value
'        .N0h = _n(_fxr(s.Key)(0)).Head
'        .N1h = _n(_fxr(s.Key)(1)).Head
'        .bFlux = _f(_fxr(s.Key)(2)).Flux
'        If _fxr(s.Key).Count = 3 Then ' ghost node boundary condition
'            '.fFlux = _f(_fxr(s.Key)(2)).Flux
'            .AvgOrthoFlux = 0.0
'        Else
'            .fFlux = _f(_fxr(s.Key)(3)).Flux
'            Dim qorth As Double = 0.0
'            For i = 4 To 7
'                qorth += _f(_fxr(s.Key)(i)).Flux
'            Next
'            .AvgOrthoFlux = qorth / 4.0 ' eq. 9/10 average orthogonal flux
'        End If
'    End With
'Next
Parallel.ForEach(_s, Sub(s)
						 With s.Value
							 .N0h = _n(_fxr(s.Key)(0)).Head
							 .N1h = _n(_fxr(s.Key)(1)).Head
							 .bFlux = _f(_fxr(s.Key)(2)).Flux
							 If _fxr(s.Key).Count = 3 Then ' ghost node boundary condition
								 '.fFlux = _f(_fxr(s.Key)(2)).Flux
								 .AvgOrthoFlux = 0.0
							 Else
								 .fFlux = _f(_fxr(s.Key)(3)).Flux
								 Dim qorth As Double = 0.0
								 For i = 4 To 7
									 qorth += _f(_fxr(s.Key)(i)).Flux
								 Next
								 .AvgOrthoFlux = qorth / 4.0 ' eq. 9/10 average orthogonal flux
							 End If
						 End With
					 End Sub)
End Sub
Private Function UpdateHeads() As Double
Dim resid As Double = 0.0, aresid As Double = 0.0, d1 = _dt / _dx '^ 2.0 ' error in equation 11, see eq. 20 in de Almeda etal 2012
For Each n In _n
	If n.Key >= _gf.nCells Then Exit For ' ghost node boundary condition
	With n.Value
		Dim dh = d1 * (_f(.FaceID(2)).Flux - _f(.FaceID(0)).Flux + _f(.FaceID(3)).Flux - _f(.FaceID(1)).Flux) ' eq 11
		If Not IsNothing(_r) AndAlso _r.ContainsKey(n.Key) Then dh += _r(n.Key)
		Dim adh = Math.Abs(dh)
		If adh > aresid Then
			aresid = adh
			resid = dh
		End If
		.Head += dh
	End With
Next
Return resid
End Function
Private Sub pUpdateHeads()
Dim d1 = _dt / _dx '^ 2.0 ' error in equation 11, see eq. 20 in de Almeda etal 2012
Parallel.ForEach(_n, Sub(n)
						 If n.Key < _gf.nCells Then ' ghost node boundary condition
							 With n.Value
								 .Head += d1 * (_f(.FaceID(2)).Flux - _f(.FaceID(0)).Flux + _f(.FaceID(3)).Flux - _f(.FaceID(1)).Flux) ' eq 11
							 End With
						 End If
					 End Sub)
If Not IsNothing(_r) Then
	Parallel.ForEach(_r, Sub(r)
							 With _n(r.Key)
								 .Head += r.Value * _dt ' m
							 End With
						 End Sub)
End If
End Sub


