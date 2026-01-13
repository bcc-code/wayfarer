# Test Data

This directory contains test images for the image processing service unit tests.

## Image Attribution

All test images in this directory were generated using ImageMagick for testing purposes:

- `test_image.jpg` - 100x80 JPEG with red background and blue rectangle
- `test_image.png` - 200x150 PNG with green background and yellow circle
- `test_image.gif` - 50x50 GIF with blue background and white rectangle
- `test_image.webp` - 120x90 WebP with purple background and orange triangle

Generated using ImageMagick commands like:
```bash
convert -size 100x80 xc:red -fill blue -draw "rectangle 30,20 70,60" test_image.jpg
```
