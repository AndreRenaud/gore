# 🔥 GORE 🔥
## A Minimal Go Port of doomgeneric

```
    ██████╗  ██████╗  ██████╗ ███╗   ███╗
    ██╔══██╗██╔═══██╗██╔═══██╗████╗ ████║
    ██║  ██║██║   ██║██║   ██║██╔████╔██║
    ██║  ██║██║   ██║██║   ██║██║╚██╔╝██║
    ██████╔╝╚██████╔╝╚██████╔╝██║ ╚═╝ ██║
    ╚═════╝  ╚═════╝  ╚═════╝ ╚═╝     ╚═╝
                    .GO
```

## 💀 WHAT FRESH HELL IS THIS?

This is a **minimal, platform-agnostic Go port** of the legendary DOOM engine, transpiled from the `doomgeneric` codebase. No CGo. No platform dependencies. Just pure, unadulterated demon-slaying action powered by the glory of Go's cross-compilation.

The original C code was converted to Go using (modernc.org/ccgo/v4), by cznic (https://gitlab.com/cznic/doomgeneric.git). This was then manually cleaned up to remove a lot of manual pointer manipulation, and make things more Go-ish, whilst still maintaining compatibility with the original Doom, and its overall structure.

## 🔫 FEATURES

- ✅ **Platform Agnostic**: Runs anywhere Go runs
- ✅ **Minimal Dependencies**: Only requires Go standard library
- ✅ **Multiple DOOM Versions**: Supports DOOM, DOOM II, Ultimate DOOM, Final DOOM
- ✅ **WAD File Support**: Bring your own demons via WAD files
- ✅ **Memory Safe**: Go's GC protects you from buffer overflows (but not from Cacodemons) (WIP - 95% complete)
- ✅ **Cross Compilation**: Build for any target from any platform

## 🚀 INSTALLATION

### Prerequisites
- Go 1.24+
- A WAD file

### Running the examples
These examples are both very minimal, and whilst technically run the game, they are not really fully complete games in their own right (ie: Missing key bindings etc...)
#### Web based
```bash
git clone <this-repo>
cd gore
go run ./example/webserver
```
Now browse to http://localhost:8080 to play

#### Ebitengine
```bash
go run ./example/ebitengine
```
The window should pop up to run Doom

### Getting WAD Files
You need the game data files (WAD) to run DOOM:
- **Shareware**: Download `doom1.wad` (free)
- **Retail**: Use your legally owned copy of DOOM.WAD or doom2.wad
- **Ultimate DOOM**: doom.wad from Ultimate DOOM
- **Final DOOM**: tnt.wad or plutonia.wad

## 🔧 PLATFORM IMPLEMENTATION

Similar to `doomgeneric`, the actual input/output is provided externally. The following interface is required:
```go
type DoomFrontend interface {
	DrawFrame(img *image.RGBA)
	SetTitle(title string)
	GetEvent(event *DoomEvent) bool
}
```

| Function | Purpose |
|----------|---------|
| `DrawFrame()` | Render the frame to your display |
| `SetTitle()` | Set the window title as appropriate to the given WAD |
| `GetEvent()` | Report key presses/mouse movements |

## 📜 LICENSE

DOOM source code is released under the GNU General Public License.  
This Go port maintains the same licensing terms.
