# vision zero speed camera

## create a project

a) new project /path/to/video
b) make a working directory for this project ymdhms_name
c) pick all the settings (web UI)
d) extract positions
e) render video

## Video Formating

If the image file is not detected peroperly:

If it's already h264

```bash
ffmpeg -i in.MOV -an -c copy out.m4a
```

Or to convert to h264

```bash
ffmpeg -i in.avi -an -c:v libx264 data/out.m4a

```

----


CGO_LDFLAGS="-L/usr/local/Cellar/ffmpeg/3.3/lib" CGO_CFLAGS="-I/usr/local/Cellar/ffmpeg/3.3/include" gb build


# Image libraries

> https://github.com/bamiaux/rez Image resizing in pure Go and SIMD
http://www.imagemagick.org/Usage/distorts/
https://godoc.org/github.com/rainycape/magick
https://godoc.org/gopkg.in/gographics/imagick.v3/imagick#DistortImageMethod
http://www.imagemagick.org/script/magick-wand.php

https://www.socketloop.com/tutorials/golang-edge-detection-with-sobel-method
https://github.com/anthonynsimon/bild

drawing libraries
https://godoc.org/golang.org/x/image/vector#Rasterizer.LineTo
https://github.com/llgcode/draw2d

# papers to read
shadow removal
http://ai2-s2-pdfs.s3.amazonaws.com/1553/ad9f771511172c943c7d4209767acaa1bc73.pdf
vanishing point
http://www.cim.mcgill.ca/~langer/558/2009/lecture14.pdf
A c++ implementation of vanishiong point 
https://github.com/yashchandak/Vanishing_Point_Detection
via example code
https://github.com/opencv/opencv/blob/master/samples/cpp/lsd_lines.cpp

lectures on vanishing point
http://www.cim.mcgill.ca/~langer/558.html
http://www.cim.mcgill.ca/~langer/558/2-translation.pdf
http://www.cim.mcgill.ca/~langer/558/15-HoughRANSAC.pdf

https://github.com/ClayFlannigan/icp/blob/master/icp.py
