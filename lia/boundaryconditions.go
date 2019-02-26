package lia



// 'Function SetGhostNodes(bFace As List(Of Integer)) As List(Of Integer)
// '    Dim lstC As New List(Of Integer)
// '    For Each f In bFace
// '        If Not _bf(f) Then Stop ' only applicable to boundary faces
// '        With _f(f)
// '            If .NodeIDs(0) = -1 And .NodeIDs(1) = -1 Then
// '                Stop ' error
// '            ElseIf .NodeIDs(0) = -1 Then
// '                lstC.Add(_n.Count)
// '                .NodeFromID = _n.Count
// '                _s.Add(f, New _state)
// '                Dim bfd = IIf(_gf.IsUpwardFace(f), 1, 0), nid = .NodeIDs(1)
// '                _fxr.Add(f, { .NodeIDs(0), nid, _n(nid).FaceID(bfd)})
// '                _n.Add(_n.Count, New _node With {.Head = _n(nid).Elevation, .Elevation = _n(nid).Elevation - 0.001, .Mannings_n = _n(nid).Mannings_n})
// '                .Initialize(_n(.NodeIDs(0)), _n(nid), _theta, _dx)
// '            ElseIf .NodeIDs(1) = -1 Then
// '                lstC.Add(_n.Count)
// '                .NodeToID = _n.Count
// '                _s.Add(f, New _state)
// '                Dim bfd = IIf(_gf.IsUpwardFace(f), 0, 2), nid = .NodeIDs(0)
// '                _fxr.Add(f, {nid, .NodeIDs(1), _n(nid).FaceID(bfd)})
// '                _n.Add(_n.Count, New _node With {.Head = _n(nid).Elevation, .Elevation = _n(nid).Elevation - 0.001, .Mannings_n = _n(nid).Mannings_n})
// '                .Initialize(_n(nid), _n(.NodeIDs(1)), _theta, _dx)
// '            Else
// '                Stop ' error
// '            End If
// '            _bf(f) = False
// '        End With
// '    Next
// '    Return lstC
// 'End Function

Function SetHeadBC(faces As List(Of Integer), Value As Double) As List(Of Integer)
	Dim dic1 As New Dictionary(Of Integer, Double)
	For Each f In faces
		dic1.Add(f, Value)
	Next
	Return Me.SetHeadBC(dic1)
End Function
Function SetHeadBC(fbc As Dictionary(Of Integer, Double)) As List(Of Integer)
	Dim lstC As New List(Of Integer)
	For Each f In fbc
		If Not _bf(f.Key) Then Stop // only applicable to boundary faces
		With _f(f.Key)
			If .NodeIDs(0) = -1 And .NodeIDs(1) = -1 Then
				Stop // error
			ElseIf .NodeIDs(0) = -1 Then
				lstC.Add(_n.Count)
				.NodeFromID = _n.Count
				_s.Add(f.Key, New _state)
				Dim bfd = IIf(_gf.IsUpwardFace(f.Key), 1, 0), nid = .NodeIDs(1)
				_fxr.Add(f.Key, { .NodeIDs(0), nid, _n(nid).FaceID(bfd)})
				_n.Add(_n.Count, New _node With {.Head = f.Value, .Elevation = _n(nid).Elevation - 0.001, .Mannings_n = _n(nid).Mannings_n}) // ghost node
				.Initialize(_n(.NodeIDs(0)), _n(nid), _theta, _dx)
			ElseIf .NodeIDs(1) = -1 Then
				lstC.Add(_n.Count)
				.NodeToID = _n.Count
				_s.Add(f.Key, New _state)
				Dim bfd = IIf(_gf.IsUpwardFace(f.Key), 0, 2), nid = .NodeIDs(0)
				_fxr.Add(f.Key, {nid, .NodeIDs(1), _n(nid).FaceID(bfd)})
				_n.Add(_n.Count, New _node With {.Head = f.Value, .Elevation = _n(nid).Elevation - 0.001, .Mannings_n = _n(nid).Mannings_n}) // ghost node
				.Initialize(_n(nid), _n(.NodeIDs(1)), _theta, _dx)
			Else
				Stop // error
			End If
			_bf(f.Key) = False
		End With
	Next
	Return lstC
End Function
Function SetFluxBC(fbc As Dictionary(Of Integer, Double)) As List(Of Integer)
	Dim lstC As New List(Of Integer)
	For Each f In fbc
		If Not _bf(f.Key) Then Stop // only applicable to boundary faces
		With _f(f.Key)
			If .NodeIDs(0) = -1 And .NodeIDs(1) = -1 Then
				Stop // error
			ElseIf .NodeIDs(0) = -1 Then
				.Flux = f.Value
			ElseIf .NodeIDs(1) = -1 Then
				.Flux = -f.Value
			Else
				Stop // error
			End If
		End With
		lstC.Add(f.Key)
	Next
	Return lstC
End Function
Function SetFluxBC(fs As List(Of Integer), Value As Double) As List(Of Integer)
	Dim dic1 As New Dictionary(Of Integer, Double)
	For Each f In fs
		dic1.Add(f, Value)
	Next
	Return Me.SetFluxBC(dic1)
End Function

Sub SetFlux(fs As List(Of Integer), Value As Double)
	For Each f In fs
		If Not _bf(f) Then Stop // only applicable to boundary faces
		With _f(f)
			If .NodeIDs(0) = -1 And .NodeIDs(1) = -1 Then
				Stop // error
			ElseIf .NodeIDs(0) = -1 Then
				.Flux = Value
			ElseIf .NodeIDs(1) = -1 Then
				.Flux = -Value
			Else
				Stop // error
			End If
		End With
	Next
End Sub

Sub SetHeads(ns As List(Of Integer), Value As Double)
	For Each n In ns
		_n(n).Head = Value
	Next
End Sub
Sub SetHeads(nh As Dictionary(Of Integer, Double))
	For Each n In nh
		_n(n.Key).Head = n.Value
	Next
End Sub
Sub SetHeads(h As Double)
	For Each n In _n.Values
		n.Head = h
	Next
End Sub
