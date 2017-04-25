
# Rotate and Crop

Adjust the rotate angle, then pick the X/Y range


```bash
julia main_rotate.jl --file=../IMG_2399_1024.MOV -r .321 -x 721 -X 1821 -y -24 -Y 201
```

http://127.0.0.1:8080/?filename=..%2F139_5th_ave_20151229.mov&rotate=1e-05&bbox=1x210+1920x1080&mask=573%3A575&mask=387%3A390&mask=&mask=&mask=
http://127.0.0.1:8080/?filename=..%2Ftraffic_speed_1080p_edited_reencoded.mp4&rotate=0.3250614737779496&bbox=607x325+1732x566&mask=96%3A98&mask=380x54+1124x2&mask=17x162+1125x239&mask=133%3A135&mask=

http://127.0.0.1:8080/?filename=..%2Ftraffic_speed_1080p_edited_reencoded.mp4&rotate=0.32410000508298165&bbox=536x326+1775x564&mask=452x51+1240x1&mask=1x166+1240x239&mask=87%3A88&mask=127%3A128&mask=1x1+277x36&min_mass=250&blur=5

http://www.nyc.gov/html/dot/html/about/datafeeds.shtml
  > Real Time speed data

http://opentraffic.io/



# extract keyframes

If the image file is not detected peroperly:
```
ffmpeg -i IMG_8491.MOV -an -c copy IMG_8491.m4a
```

CGO_LDFLAGS="-L/usr/local/Cellar/ffmpeg/3.3/lib" CGO_CFLAGS="-I/usr/local/Cellar/ffmpeg/3.3/include" gb build


# Image libraries

> https://github.com/bamiaux/rez Image resizing in pure Go and SIMD
http://www.imagemagick.org/Usage/distorts/
https://godoc.org/github.com/rainycape/magick
https://godoc.org/gopkg.in/gographics/imagick.v3/imagick#DistortImageMethod