#!/bin/bash
set -e
curl --silent 'https://raw.githubusercontent.com/fiji/fiji/master/luts/glasbey.lut' | tail -256 | awk 'BEGIN {print "package labelimg\n// https://raw.githubusercontent.com/fiji/fiji/master/luts/glasbey.lut\n\nimport \"image/color\"\n var Glasbey = []color.Color{"} {print "color.RGBA{"$2","$3","$4",0},"} END{print "}"}' > glasbey.go
go fmt