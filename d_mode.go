package gore

type gamemission_t int32

const doom gamemission_t = 0
const doom2 gamemission_t = 1
const pack_tnt gamemission_t = 2
const pack_plut gamemission_t = 3
const pack_chex gamemission_t = 4
const pack_hacx gamemission_t = 5
const heretic gamemission_t = 6
const hexen gamemission_t = 7
const strife gamemission_t = 8
const none gamemission_t = 9

type gamemode_t int32

const shareware gamemode_t = 0
const registered gamemode_t = 1
const commercial gamemode_t = 2
const retail gamemode_t = 3
const indetermined gamemode_t = 4

type gameversion_t int32

const exe_doom_1_2 gameversion_t = 0
const exe_doom_1_666 gameversion_t = 1
const exe_doom_1_7 gameversion_t = 2
const exe_doom_1_8 gameversion_t = 3
const exe_doom_1_9 gameversion_t = 4
const exe_hacx gameversion_t = 5
const exe_ultimate gameversion_t = 6
const exe_final gameversion_t = 7
const exe_final2 gameversion_t = 8
const exe_chex gameversion_t = 9

type skill_t int32

const sk_baby skill_t = 0
const sk_easy skill_t = 1
const sk_medium skill_t = 2
const sk_nightmare skill_t = 4

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
