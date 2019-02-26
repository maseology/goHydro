package lia

// LIA local inertial approximation to the 2D SWEs
// ref: de Almeida, G.A.M., P. Bates, 2013. Applicability of the local intertial approximation of the shallow water equations to flood modeling. Water Resources Research 49: 4833-4844.
// see also: de Almeida Bates Freer Souvignet 2012 Improving the stability of a simple formulation of the shallow water equations for 2-D flood modeling
//           Sampson etal 2012 An automated routing methodology to enable direct rainfall in high resolution shallow water models
// similar in theory to LISFLOOD-FP
type LIA struct {
 f map[int]face
 n map[int]node
 s map[int]state
 r map[int]float64 // vertical influx [m/s]
 fxr map[int][]int
 bf map[int]bool
 tacum, dt, dx, alpha, theta, tresid float64
}

 type state struct {
     n0h,n1h,fflux,bflux,avgOrthoFlux float64
 }
 type node struct {
z,h,n float64 // elevation, head, manning's n
fid []int // face id
 }
type face struct {
forth []int
q, t, dx, n2, zx float64 // parameters and variables: q: flux; t: theta; 
    nfrom, nto, ffw, fbw int // node and face identifiers
}

func newFace() face {
    var f face
    f.nfrom=-1
    f.nto=-1
    f.ffw=-1
    f.fbw=-1
}

func (f *face) nodeIDs() float64, float64 {
    return f.nfrom,f.nto // from node id, to node id
}

func (f *face) isBoundary() bool {
    return len(f.forth)==0
}

func (f *face) IsInactive bool {
    return nfrom==-1 && nto==-1
}


            Public ReadOnly Property IdColl As Integer()
                Get
                    Dim in1(7) As Integer
                    in1(0) = _nfrom
                    in1(1) = _nto
                    in1(2) = _fbw
                    in1(3) = _ffw
                    For i = 0 To 3
                        in1(4 + i) = _forth(i)
                    Next
                    Return in1
                End Get
            End Property

            Sub New()
            End Sub
            Sub New(GF As Grid.Face, fid As Integer)
                With GF
                    _nfrom = .FaceCell(fid)(0)
                    _nto = .FaceCell(fid)(1)
                    If _nfrom = -1 Or _nto = -1 Then
                        _q = 0 ' (default) no flow boundary
                    Else
                        ReDim _forth(3) ' orthogonal faces
                        If .IsUpwardFace(fid) Then ' upward meaning direction normal to face
                            _ffw = .CellFace(_nto)(1)
                            _fbw = .CellFace(_nfrom)(3)
                            _forth(0) = .CellFace(_nfrom)(2)
                            _forth(1) = .CellFace(_nfrom)(0)
                            _forth(2) = .CellFace(_nto)(2)
                            _forth(3) = .CellFace(_nto)(0)
                        Else
                            _ffw = .CellFace(_nto)(0)
                            _fbw = .CellFace(_nfrom)(2)
                            _forth(0) = .CellFace(_nfrom)(3)
                            _forth(1) = .CellFace(_nfrom)(1)
                            _forth(2) = .CellFace(_nto)(3)
                            _forth(3) = .CellFace(_nto)(1)
                        End If
                    End If
                End With
            End Sub

            Sub Initialize(Node0 As _node, Node1 As _node, Theta As Double, cellsize As Double)
                _t = Theta
                _dx = cellsize
                _zx = Math.Max(Node0.Elevation, Node1.Elevation)
                _n2 = ((Node0.Mannings_n + Node1.Mannings_n) / 2.0) ^ 2.0
            End Sub

            Public Sub UpdateFlux(s As _state, dt As Double)
                With s
                    Dim hf = Math.Max(.N0h, .N1h) - _zx
                    If hf <= 0.000001 Then
                        _q = 0.0
                    Else
                        Dim qmag = Math.Sqrt(_q ^ 2.0 + .AvgOrthoFlux ^ 2.0) ' eq. 8
                        'Dim qmag = Math.Abs(_q) ' de Almeda etal 2012
                        _q = _t * _q + 0.5 * (1.0 - _t) * (.fFlux + .bFlux) - 9.80665 * hf * dt * (.N1h - .N0h) / _dx ' eq. 7 numer
                        _q /= 1 + 9.80665 * dt * _n2 * qmag / hf ^ 2.33333 ' eq.7 denom
                    End If
                End With
            End Sub

        End Class (face)
