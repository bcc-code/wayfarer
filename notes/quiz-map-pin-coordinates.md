# MAP_PIN coordinate system

A `MAP_PIN` quiz question asks a participant to drop a pin on a top-down map
image. All of its geometry is stored in **the natural pixel space of the map
image**: the correct point and the scoring radii (`POINT_RADIUS` mode), the
vertices of the correct area (`POLYGON` mode), and every submitted pin.

- origin `(0, 0)` is the top-left corner of the image
- `x` grows to the right, `y` grows down
- one unit is one pixel of the image file as uploaded (its intrinsic size,
  `naturalWidth` × `naturalHeight` in the browser) — not a CSS pixel and not a
  screen pixel
- values are floats; a pin is allowed anywhere in `[0, mapWidth] × [0, mapHeight]`

The question stores `mapWidth` / `mapHeight` alongside the geometry. They
describe the image the target or polygon was placed on.

See `quiz-map-pin-question.md` for the feature as a whole (scoring,
visibility, results).

## 1. Why natural pixels and not normalized coordinates

The alternative was normalized coordinates, `x, y ∈ [0, 1]` relative to the
image's width and height. We chose pixels because:

- **Distance is honest without extra data.** `POINT_RADIUS` scoring is
  `hypot(x − tx, y − ty)` compared against two radii. In pixel space one unit
  is the same length horizontally and vertically, so that is plain Euclidean
  distance. In normalized space one unit along `x` is the image width and one
  unit along `y` is the image height; on a non-square map a circle becomes an
  ellipse unless every distance calculation also knows the aspect ratio.
  Normalized coordinates would therefore need the aspect ratio stored on the
  question anyway — at which point storing width and height is simpler.
  (Point-in-polygon would work in either space, since scaling an axis keeps
  inside inside; the boundary tolerance would not, so one frame for both
  modes keeps it simple.)
- **Grading does not depend on upload metadata.** `file_uploads` does record
  width/height, but they can be null (URLs not uploaded through `/upload`,
  decode failures) and the Go decoder ignores EXIF orientation, so a rotated
  JPEG can report swapped dimensions. Grading reads only values stored on the
  question.
- **Readable and debuggable.** "Target at (812, 430), full points within 40,
  none beyond 200" can be checked against the image in any image viewer. The
  radii mean something an organizer can picture.

The cost is described in section 3.

## 2. How the frontend handles zoom and pan

Zoom and pan are purely a rendering concern. The backend never sees them.

The map is drawn with a transform from natural pixels to screen pixels:

```
screen = natural × scale + offset
scale  = (renderedWidth / naturalWidth) × zoom
```

`renderedWidth / naturalWidth` is how much the layout shrinks or enlarges the
image to fit; `zoom` and `offset` come from the pinch/scroll/drag state.

- **Placing a pin:** take the tap or click position relative to the map
  element and invert the transform: `natural = (screen − offset) / scale`.
  Clamp the result to `[0, naturalWidth] × [0, naturalHeight]` and send that
  value as `mapPosition`.
- **Drawing pins and shapes** (the user's own pin, the correct point or
  polygon after reveal, pins on the results screen): apply the forward
  transform to every stored natural coordinate — for a polygon, to each
  vertex. Redraw on every zoom/pan/resize; the stored values never change.
  Radii scale by `scale` alone (no offset).
- The admin editor works the same way: the organizer clicks the target, or
  clicks the polygon's vertices in order, the editor stores the natural
  coordinates, and fills `mapWidth` / `mapHeight` from the image's
  `naturalWidth` / `naturalHeight`.

Never send screen coordinates, coordinates scaled by the current zoom, or
values in CSS pixels.

## 3. Why changing the map's resolution needs reconfiguration

The target, both radii and every polygon vertex are numbers in the pixel space
of one particular image file. If the map is replaced with a file of a
different resolution, those numbers keep their values but now refer to
different places and lengths:

- Re-exporting the map at twice the size: the stored target `(812, 430)` now
  points at a spot in the upper-left quarter of the old location, and a 40 px
  full-points radius covers half the real distance it used to. A polygon
  shrinks to a quarter of its area, drawn around the wrong room.
- Cropping or padding the map shifts every position, even at the same scale.

So **after swapping the map for a file with different dimensions or framing,
the organizer must place the target or redraw the polygon again, and re-check
the radii.** The backend rejects a target, vertex or pin outside
`[0, mapWidth] × [0, mapHeight]`, and the admin UI can warn when the stored
`mapWidth` / `mapHeight` disagree with the loaded image's natural size.
Replacing the file with one of identical dimensions and framing (for example a
colour correction) needs nothing.

Do not change the map once anyone has answered: pins already submitted were
placed on the old image. (The backend locks the scoring config of a question
with responses, but not its image.)

## 4. Changing only the displayed size does not affect scoring

How big the map appears — phone vs. tablet vs. the projector, CSS layout,
device pixel ratio, zoom level — only changes `scale` and `offset` in the
transform from section 2. The inverse transform removes them again, so the
same physical spot on the map always yields the same natural coordinates, and
therefore the same distance and points, on every device.
