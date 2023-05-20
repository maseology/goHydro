package tem

/*



   Console.WriteLine(" fixing flat regions")
            'Dim fpmsv = _fpm
            'If fpmsv <> FlowpathMethod.MultiDirectionDecending Then
            '    _fpm = FlowpathMethod.MultiDirectionDecending 'helpful to go to a D8
            '    Me.ReBuildFlowpaths()
            'End If
            If IsNothing(_cascid) Then Me.BuildCascadeIndex()

            ' collect flat regions
            Console.WriteLine("  collecting flat regions")
            Dim sc0 = New Crawler(Me.Elevations)
            Dim dicZ As New Dictionary(Of Integer, Double)
            With _gd
                For Each rc In .ActivesRowCol
                    Dim i = rc.Row, j = rc.Col
                    If _dem(i)(j).Elevation = -9999.0 Then Continue For
                    Dim cid = .CellID0(i, j)
                    If dicZ.ContainsKey(cid) Then Continue For
                    If _cascid(i)(j) < 0 Then Continue For
                    If _dem(i)(j).Gradient = 0.0 Or _cascid(i)(j) = 5 Then
                        For Each c In sc0.CollectNeighbouringCells(cid)
                            'With .CellIDToRowCol(c)
                            '    If _cascid(.Row, .Col) < 0 Then
                            '        _dem(.Row, .Col).Elevation -= _inc
                            '    Else
                            '        dicZ.Add(c, _dem(i, j).Elevation)
                            '    End If
                            'End With
                            If Not .IsActive(i, j) Then Stop
                            If Not .IsActive(c) Then Stop
                            If _dem(i)(j).Elevation = -9999.0 Then Stop
                            dicZ.Add(c, _dem(i)(j).Elevation)
                        Next
                    End If
                Next
                'For i = 0 To .NumRows - 1
                '    For j = 0 To .NumCols - 1
                '        If _dem(i)(j).Elevation = -9999.0 Then Continue For
                '        Dim cid = .CellID0(i, j)
                '        If dicZ.ContainsKey(cid) Then Continue For
                '        If _cascid(i)(j) < 0 Then Continue For
                '        If _dem(i)(j).Gradient = 0.0 Or _cascid(i)(j) = 5 Then
                '            For Each c In sc0.CollectNeighbouringCells(cid)
                '                'With .CellIDToRowCol(c)
                '                '    If _cascid(.Row, .Col) < 0 Then
                '                '        _dem(.Row, .Col).Elevation -= _inc
                '                '    Else
                '                '        dicZ.Add(c, _dem(i, j).Elevation)
                '                '    End If
                '                'End With
                '                dicZ.Add(c, _dem(i)(j).Elevation)
                '            Next
                '        End If
                '    Next
                'Next
            End With

            Console.WriteLine("  fixing flat regions")
            If dicZ.Count = 0 Then Exit Sub
            Dim sc1 As New Crawler(dicZ, _gd), buf = sc1.GroupBufferCells, dicGwBuff As New Dictionary(Of Integer, FlatRegion) ', inFR = _gd.NullArray(-9999)
            For Each g In sc1.Group
                Dim dicFR As New Dictionary(Of Integer, Double)
                For Each c In g.Value
                    With _gd.CellIDToRowCol(c)
                        If _dem(.Row)(.Col).Elevation = -9999.0 Then Stop
                        dicFR.Add(c, _dem(.Row)(.Col).Elevation)
                    End With
                Next
                If dicFR.Count <= 1 Then
                    Dim dem0 = Double.MaxValue
                    For Each c In buf(g.Key)
                        With _gd.CellIDToRowCol(c)
                            If _dem(.Row)(.Col).Elevation = -9999.0 Then Continue For
                            If _dem(.Row)(.Col).Elevation < dem0 Then dem0 = _dem(.Row)(.Col).Elevation
                        End With
                    Next
                    With _gd.CellIDToRowCol(g.Value.First)
                        If _dem(.Row)(.Col).Elevation <= dem0 Then _dem(.Row)(.Col).Elevation = dem0 + 0.00001
                    End With
                Else
                    Dim dicB As New Dictionary(Of Integer, Double)
                    For Each c In buf(g.Key)
                        With _gd.CellIDToRowCol(c)
                            If _dem(.Row)(.Col).Elevation = -9999.0 Then Continue For
                            dicB.Add(c, _dem(.Row)(.Col).Elevation)
                        End With
                    Next
                    dicGwBuff.Add(g.Key, New FlatRegion(dicFR, dicB, _gd))
                End If
                'For Each c In g.Value
                '    With _gd.CellIDToRowCol(c)
                '        inFR(.Row)(.Col) = g.Key
                '    End With
                'Next
            Next
            'mmIO.ArrayToBinary("M:\flatregions.indx", inFR)
            For Each f In dicGwBuff
                'Console.Write(f.Key & " .")
                If f.Value.HasOutlet Then f.Value.GradientTowardLower()
                'Console.Write("^")
                f.Value.GradientFromUpper()
                For Each z In f.Value.NewElevations
                    With _gd.CellIDToRowCol(z.Key)
                        _dem(.Row)(.Col).Elevation = z.Value
                    End With
                Next
            Next
            'Parallel.ForEach(dicGwBuff.Values, Sub(f)
            '                                       If f.HasOutlet Then f.GradientTowardLower()
            '                                       f.GradientFromUpper()
            '                                       For Each z In f.NewElevations
            '                                           With _gd.CellIDToRowCol(z.Key)
            '                                               _dem(.Row)(.Col).Elevation = z.Value
            '                                           End With
            '                                       Next
            '                                   End Sub)

            For Each f In dicGwBuff.Values
                For Each z In f.NewElevations
                    With _gd.CellIDToRowCol(z.Key)
                        _dem(.Row)(.Col).Elevation = z.Value
                    End With
                Next
            Next

            Console.WriteLine("  re-building flowpaths")
            'If fpmsv <> FlowpathMethod.MultiDirectionDecending Then _fpm = fpmsv
            Me.ReBuildFlowpaths()
            If Not IsNothing(_cascid) Then
                Me.AddSinks()
                Me.BuildCascadeIndex()
            End If




*/
