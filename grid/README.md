# goHydro.grid

Used for gridded data. Reads from and Writes to [band interleaved by line (*.BIL) raster files](https://desktop.arcgis.com/en/arcmap/10.5/manage-data/raster-and-images/bil-bip-and-bsq-raster-files.htm).



### Structures

* **`definition`** basic grid metadata (origin, number of rows/columns, rotation, cell widths, etc.)
* **`face`** is an alternate grid organization scheme based on the shared faces among grid cells.
* **`indx`** for grids of integer data, generally assumed to have categorical implications.
* **`intersect`** functions needed to rescale data from differing grid `definitions`.
* **`real`** for grids of real (floating point) data.
* **`sws`** specialized tools using the above functions to define set of topologically ordered sub-watersheds.
