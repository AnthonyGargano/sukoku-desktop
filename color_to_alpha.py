"""
GIMP Color to Alpha — Python implementation.
Removes a background color from an image by making it transparent,
handling semi-transparent edge pixels correctly (same as GIMP's algorithm).
"""
import sys
from PIL import Image

def color_to_alpha(image, color=(255, 255, 255)):
    """
    Exact port of GIMP's plug-in-colortoalpha algorithm.
    color = (R, G, B) of the background color to remove.
    Returns a new RGBA image with that color made transparent.
    """
    image = image.convert("RGBA")
    pixels = list(image.getdata())
    cr, cg, cb = float(color[0]), float(color[1]), float(color[2])
    new_pixels = []

    for (r, g, b, a) in pixels:
        r, g, b = float(r), float(g), float(b)

        # For each channel, calculate how much "background color" this pixel contains
        def chan_alpha(chan, col):
            if col == 255.0:
                return (255.0 - chan) / 255.0
            elif col == 0.0:
                return chan / 255.0
            elif chan <= col:
                return (col - chan) / col
            else:
                return (chan - col) / (255.0 - col)

        ar = chan_alpha(r, cr)
        ag = chan_alpha(g, cg)
        ab = chan_alpha(b, cb)

        # Overall alpha is the maximum across all channels
        alpha = max(ar, ag, ab)

        if alpha < 1.0 / 255.0:
            # Fully transparent — keep colour but zero alpha
            new_pixels.append((int(r), int(g), int(b), 0))
        else:
            # Solve: pixel = alpha * new_col + (1-alpha) * bg_col
            # => new_col = (pixel - bg_col) / alpha + bg_col
            def new_chan(chan, col):
                return min(255, max(0, int(round((chan - col) / alpha + col))))

            nr = new_chan(r, cr)
            ng = new_chan(g, cg)
            nb = new_chan(b, cb)
            new_pixels.append((nr, ng, nb, int(round(alpha * 255.0))))

    result = image.copy()
    result.putdata(new_pixels)
    return result


if __name__ == "__main__":
    src = sys.argv[1] if len(sys.argv) > 1 else "Icon.png"
    dst = sys.argv[2] if len(sys.argv) > 2 else "Icon_transparent.png"

    # Background color to remove: white = (255,255,255) from the JPEG artefact
    bg = (255, 255, 255)

    print(f"Loading: {src}")
    img = Image.open(src)
    print(f"Mode: {img.mode}, Size: {img.size}")

    print(f"Running Color-to-Alpha (removing white #{bg[0]:02X}{bg[1]:02X}{bg[2]:02X})...")
    result = color_to_alpha(img, bg)

    print(f"Saving: {dst}")
    result.save(dst, "PNG")
    print("Done.")
