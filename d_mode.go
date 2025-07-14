package gore

func d_GameMissionString(mission gamemission_t) string {
	switch mission {
	case none:
		fallthrough
	default:
		return "none"
	case doom:
		return "doom"
	case doom2:
		return "doom2"
	case pack_tnt:
		return "tnt"
	case pack_plut:
		return "plutonia"
	case pack_hacx:
		return "hacx"
	case pack_chex:
		return "chex"
	case heretic:
		return "heretic"
	case hexen:
		return "hexen"
	case strife:
		return "strife"
	}
	return ""
}
