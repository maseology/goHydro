

Dim gd1 As New Grid.Definition("test", 5, 4, 100.0)
Dim mz As New Dictionary(Of Integer, Double)
Dim mh As New Dictionary(Of Integer, Double)
Dim mn As New Dictionary(Of Integer, Double)
Dim cnt = 0
For i = 0 To gd1.NumRows - 1
    For j = 0 To gd1.NumCols - 1
        mz.Add(cnt, 12.0 - CDbl(i * j))
        mh.Add(cnt, mz(cnt) + 1.0)
        mn.Add(cnt, 0.05)
        cnt += 1
    Next
Next

Console.WriteLine()
cnt = 0
For i = 0 To gd1.NumRows - 1
    For j = 0 To gd1.NumCols - 1
        Console.Write("{0,10:0.000}", mz(cnt))
        cnt += 1
    Next
    Console.WriteLine()
Next

Dim lia As New Hydraulic.LocalInertialSWE(gd1, mz, mh, mn)
Dim mout = lia.Solve(86400)

Console.WriteLine()
cnt = 0
Dim wbal = 0.0
For i = 0 To gd1.NumRows - 1
    For j = 0 To gd1.NumCols - 1
        Console.Write("{0,10:0.000}", mout(cnt))
        wbal += mout(cnt) - mz(cnt)
        cnt += 1
    Next
    Console.WriteLine()
Next
Console.WriteLine(wbal)
