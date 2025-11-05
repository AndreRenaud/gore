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

## TLDR
Tired of reading already?
```bash
wget https://distro.ibiblio.org/slitaz/sources/packages/d/doom1.wad
go run github.com/AndreRenaud/gore/example/termdoom@latest
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

### Missing Features
- One instance per process: Still has a lot of the original global variables, which prevent multiple instances from running
- Random exported consts: The original C code used the standard convention of all upper case for const/enum values. This results in the Go code assuming these are exported values, when really they're internal state info
- Nice external API for state inspection: It would be good to be able to change the running state externally, without exposing everything in such a raw way
- `unsafe`: There are still some instances of `unsafe` in the code. It would be good to get rid of these to have better bounds access guarantees

## 🚀 INSTALLATION

### Prerequisites
- Go 1.24+
- A WAD file

### Running the examples
These examples are both very minimal, and whilst technically run the game, they are not really fully complete games in their own right (ie: Missing key bindings etc...). They all assume that a Doom wad is available in the current directory. The shareware Doom wad is available at https://www.doomworld.com/classicdoom/info/shareware.php, or bring your own from a commercial copy.

```bash
git clone https://github.com/AndreRenaud/gore
cd gore
```

#### Terminal based
This example renders the Doom output using ANSI color codes suitable for a 256-bit color capable terminal. It has very limited input support, as terminals typically do not support key-up events, or control-key support. So `fire` has been remapped to `,`, and it is necessary to repeatedly tap keys to get them to continue, as opposed to press & hold.
```bash
go run ./example/termdoom -iwad doom1.wad
```

<video width="640" src="https://github.com/user-attachments/assets/c461e38f-5948-4485-bf84-7b6982580a4e"></video>

#### Web based
```bash
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
    CacheSound(name string, data []byte)
    PlaySound(name string, channel, vol, sep int)
}
```

| Function | Purpose |
|----------|---------|
| `DrawFrame()` | Render the frame to your display |
| `SetTitle()` | Set the window title as appropriate to the given WAD |
| `GetEvent()` | Report key presses/mouse movements |
| `CacheSound()` | This will supply sound effect 8-bit 11025Hz mono audio samples |
| `PlaySound()` | Play a given sound effect |

Only `DrawFrame` and `GetEvent` are vital to implement to get a functioning game. The others can be left blank, and things will still basically function fine.

## 📜 LICENSE

DOOM source code is released under the GNU General Public License.  
This Go port maintains the same licensing terms.
