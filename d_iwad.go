package gore

import (
	"os"
	"strings"
)

const MAX_IWAD_DIRS = 128

type iwad_t struct {
	Fname        string
	Fmission     gamemission_t
	Fmode        gamemode_t
	Fdescription string
}

//
// This is used to get the local FILE:LINE info from CPP
// prior to really call the function in question.
//

var iwads = [14]iwad_t{
	0: {
		Fname:        "doom2.wad",
		Fmission:     doom2,
		Fmode:        commercial,
		Fdescription: "Doom II",
	},
	1: {
		Fname:        "plutonia.wad",
		Fmission:     pack_plut,
		Fmode:        commercial,
		Fdescription: "Final Doom: Plutonia Experiment",
	},
	2: {
		Fname:        "tnt.wad",
		Fmission:     pack_tnt,
		Fmode:        commercial,
		Fdescription: "Final Doom: TNT: Evilution",
	},
	3: {
		Fname:        "doom.wad",
		Fmode:        retail,
		Fdescription: "Doom",
	},
	4: {
		Fname:        "doom1.wad",
		Fdescription: "Doom Shareware",
	},
	5: {
		Fname:        "chex.wad",
		Fmission:     pack_chex,
		Fdescription: "Chex Quest",
	},
	6: {
		Fname:        "hacx.wad",
		Fmission:     pack_hacx,
		Fmode:        commercial,
		Fdescription: "Hacx",
	},
	7: {
		Fname:        "freedm.wad",
		Fmission:     doom2,
		Fmode:        commercial,
		Fdescription: "FreeDM",
	},
	8: {
		Fname:        "freedoom2.wad",
		Fmission:     doom2,
		Fmode:        commercial,
		Fdescription: "Freedoom: Phase 2",
	},
	9: {
		Fname:        "freedoom1.wad",
		Fmode:        retail,
		Fdescription: "Freedoom: Phase 1",
	},
	10: {
		Fname:        "heretic.wad",
		Fmission:     heretic,
		Fmode:        retail,
		Fdescription: "Heretic",
	},
	11: {
		Fname:        "heretic1.wad",
		Fmission:     heretic,
		Fdescription: "Heretic Shareware",
	},
	12: {
		Fname:        "hexen.wad",
		Fmission:     hexen,
		Fmode:        commercial,
		Fdescription: "Hexen",
	},
	13: {
		Fname:        "strife1.wad",
		Fmission:     strife,
		Fmode:        commercial,
		Fdescription: "Strife",
	},
}

// Array of locations to search for IWAD files
//
// "128 IWAD search directories should be enough for anybody".

var iwad_dirs [128]string
var num_iwad_dirs int32 = 0

func addIWADDir(dir string) {
	if num_iwad_dirs < MAX_IWAD_DIRS {
		iwad_dirs[num_iwad_dirs] = dir
		num_iwad_dirs++
	}
}

// This is Windows-specific code that automatically finds the location
// of installed IWAD files.  The registry is inspected to find special
// keys installed by the Windows installers for various CD versions
// of Doom.  From these keys we can deduce where to find an IWAD.

// Returns true if the specified path is a path to a file
// of the specified name.

func dirIsFile(path string, filename string) boolean {
	if strings.HasPrefix(filename, path) && path[len(path)-1] == '/' {
		return 1
	}
	return 0
}

// Check if the specified directory contains the specified IWAD
// file, returning the full path to the IWAD if found, or NULL
// if not found.

func checkDirectoryHasIWAD(dir string, iwadname string) string {
	var filename string
	// As a special case, the "directory" may refer directly to an
	// IWAD file if the path comes from DOOMWADDIR or DOOMWADPATH.
	if dirIsFile(dir, iwadname) != 0 && m_FileExists(dir) != 0 {
		return dir
	}
	// Construct the full path to the IWAD if it is located in
	// this directory, and check if it exists.
	if dir == "." {
		filename = iwadname
	} else {
		filename = dir + "/" + iwadname
	}
	fprintf_ccgo(os.Stdout, "Trying IWAD file:%s\n", filename)
	if m_FileExists(filename) != 0 {
		return filename
	}
	return ""
}

// Search a directory to try to find an IWAD
// Returns the location of the IWAD if found, otherwise NULL.

func searchDirectoryForIWAD(dir string, mask int32, mission *gamemission_t) string {
	var filename string
	var i uint64
	i = 0
	for {
		if i >= uint64(len(iwads)) {
			break
		}
		if 1<<iwads[i].Fmission&mask == 0 {
			goto _1
		}
		filename = checkDirectoryHasIWAD(dir, iwads[i].Fname)
		if filename != "" {
			*mission = iwads[i].Fmission
			return filename
		}
		goto _1
	_1:
		;
		i++
	}
	return ""
}

// When given an IWAD with the '-iwad' parameter,
// attempt to identify it by its name.

func identifyIWADByName(name string, mask int32) gamemission_t {
	var i uint64
	var mission gamemission_t
	mission = none
	i = 0
	for {
		if i >= uint64(len(iwads)) {
			break
		}
		// Check if the filename is this IWAD name.
		// Only use supported missions:
		if 1<<iwads[i].Fmission&mask == 0 {
			goto _1
		}
		// Check if it ends in this IWAD name.
		if name == iwads[i].Fname {
			mission = iwads[i].Fmission
			break
		}
		goto _1
	_1:
		;
		i++
	}
	return mission
}

//
// Build a list of IWAD files
//

func buildIWADDirList() {
	addIWADDir(".")
}

//
// Searches WAD search paths for an WAD with a specific filename.
//

func d_FindWADByName(name string) string {
	var i int32
	// Absolute path?
	if m_FileExists(name) != 0 {
		return name
	}
	buildIWADDirList()
	// Search through all IWAD paths for a file with the given name.
	i = 0
	for {
		if i >= num_iwad_dirs {
			break
		}
		// As a special case, if this is in DOOMWADDIR or DOOMWADPATH,
		// the "directory" may actually refer directly to an IWAD
		// file.
		if dirIsFile(iwad_dirs[i], name) != 0 && m_FileExists(iwad_dirs[i]) != 0 {
			return iwad_dirs[i]
		}
		// Construct a string for the full path
		path := iwad_dirs[i] + "/"
		if m_FileExists(path) != 0 {
			return path
		}
		goto _1
	_1:
		;
		i++
	}
	// File not found
	return ""
}

//
// D_TryWADByName
//
// Searches for a WAD by its filename, or passes through the filename
// if not found.
//

func d_TryFindWADByName(filename string) string {
	result := d_FindWADByName(filename)
	if result != "" {
		return result
	} else {
		return filename
	}
}

//
// FindIWAD
// Checks availability of IWAD files by name,
// to determine whether registered/commercial features
// should be executed (notably loading PWADs).
//

func d_FindIWAD(mask int32, mission *gamemission_t) string {
	var i, iwadparm int32
	var result string
	var iwadfile string
	// Check for the -iwad parameter
	//!
	// Specify an IWAD file to use.
	//
	// @arg <file>
	//
	iwadparm = m_CheckParmWithArgs("-iwad", 1)
	if iwadparm != 0 {
		// Search through IWAD dirs for an IWAD with the given name.
		iwadfile = myargs[iwadparm+1]
		result = d_FindWADByName(iwadfile)
		if result == "" {
			i_Error("IWAD file '%s' not found!", iwadfile)
		}
		*mission = identifyIWADByName(result, mask)
	} else {
		// Search through the list and look for an IWAD
		fprintf_ccgo(os.Stdout, "-iwad not specified, trying a few iwad names\n")
		result = ""
		buildIWADDirList()
		i = 0
		for {
			if !(result == "" && i < num_iwad_dirs) {
				break
			}
			result = searchDirectoryForIWAD(iwad_dirs[i], mask, mission)
			goto _1
		_1:
			;
			i++
		}
	}
	return result
}

//
// Get the IWAD name used for savegames.
//

func d_SaveGameIWADName(gamemission gamemission_t) string {
	var i uint64
	// Determine the IWAD name to use for savegames.
	// This determines the directory the savegame files get put into.
	//
	// Note that we match on gamemission rather than on IWAD name.
	// This ensures that doom1.wad and doom.wad saves are stored
	// in the same place.
	i = 0
	for {
		if i >= uint64(len(iwads)) {
			break
		}
		if gamemission == iwads[i].Fmission {
			return iwads[i].Fname
		}
		goto _1
	_1:
		;
		i++
	}
	// Default fallback:
	return "unknown.wad"
}

func d_SuggestGameName(mission gamemission_t, mode gamemode_t) string {
	for i := 0; i < len(iwads); i++ {
		if iwads[i].Fmission == mission && (mode == indetermined || iwads[i].Fmode == mode) {
			return iwads[i].Fdescription
		}
	}
	return "Unknown game?"
}
