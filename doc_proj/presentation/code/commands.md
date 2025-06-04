# Commands

## Replay

./advocate record -path ~/Obsidian/Bibliothek/Uni/Advocate/MasterProjectPresentation/code/replay/ -exec TestReplay -output -noInfo

./advocate replay -path ~/Obsidian/Bibliothek/Uni/Advocate/MasterProjectPresentation/code/replay/ -exec TestReplay -output -trace ~/Obsidian/Bibliothek/Uni/Advocate/MasterProjectPresentation/code/replay/replayOrdered -noInfo

## Select

./advocate record -path ~/Obsidian/Bibliothek/Uni/Advocate/MasterProjectPresentation/code/replay/ -exec TestSelect -output -noInfo

./advocate replay -path ~/Obsidian/Bibliothek/Uni/Advocate/MasterProjectPresentation/code/replay/ -exec TestSelect -output -trace ~/Obsidian/Bibliothek/Uni/Advocate/MasterProjectPresentation/code/replay/selectC -noInfo

## GFuzz

./advocate fuzzing -path ~/Obsidian/Bibliothek/Uni/Advocate/MasterProjectPresentation/code/GFuzz/main.go -main -fuzzingMode GFuzz -timeoutRec 10 -prog Pie -maxFuzzingRun 200

## GoPie

./advocate fuzzing -path ~/Obsidian/Bibliothek/Uni/Advocate/MasterProjectPresentation/code/goPie/main.go -main -fuzzingMode GoPie+ -timeoutRec 10 -prog Pie -maxFuzzingRun 200