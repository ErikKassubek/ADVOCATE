# New
The creation of an relevant object is recorded.
For now only the creation of channels with make is
recorded. Other creations may be added later

## Trace element
The basic form of the trace element is
```
V,[tpost],[id],[elemType],[num],[pos]
```
where `N` identifies the element as a new element. The other fields are
set as follows:
- [tpost]$\in \mathbb N$: This is the value of the global counter when the channel has finished its operation.
- [id]$\in \mathbb N$: This shows the unique id of the object.
- [elemType]: This shows the type of object created. For now only channels are recorded. But values for other types have already been impemented. The possible values are:
	- `A`: atomic variable
	- `C`: channel
	- `D`: conditional variable
	- `M`: mutex
	- `O`: once
	- `W`: wait group
- [num]: Additional field for number with the following meanings:
 - For channel: qSize
 - Else: 0
- [pos]: The last field show the position in the code, where the mutex operation
was executed. It consists of the file and line number separated by a colon (:)
