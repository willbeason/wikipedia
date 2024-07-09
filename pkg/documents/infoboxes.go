package documents

import (
	"regexp"
	"strings"
)

type InfoboxChecker struct {
	r *regexp.Regexp
}

func NewInfoboxChecker(want []string) (*InfoboxChecker, error) {
	infoboxString := strings.Join(want, "|")

	validInfoboxes, err := regexp.Compile("infobox (" + infoboxString + ")\n")
	if err != nil {
		return nil, err
	}

	return &InfoboxChecker{r: validInfoboxes}, nil
}

func (r *InfoboxChecker) Matches(rawText string) bool {
	return r.r.MatchString(strings.ToLower(rawText))
}

func PersonInfoboxes() []string {
	return []string{
		"person",
		"football biography",
		"officeholder",
		//"musical artist", // incorrectly used for bands
		"sportsperson",
		"writer",
		"scientist",
		"military person",
		"cricketer",
		"baseball biography",
		"artist",
		"basketball biography",
		"nfl biography",
		"ice hockey player",
		"christian leader",
		"rugby biography",
		"afl biography",
		"college coach",
		"academic",
		"tennis biography",
		"rugby league biography",
		"swimmer",
		"nfl player",
		"boxer",
		"saint",
		"golfer",
		"noble",
		"figure skater",
		"criminal",
		"professional wrestler",
		"diocese",
		"volleyball biography",
		"racing driver",
		"comics creator",
		"gymnast",
		"gaa player",
		"religious biography",
		"politician",
		"athlete",
		"badminton player",
		"philosopher",
		"chess player",
		"gridiron football person",
		"pageant titleholder",
		"curler",
		"nobility",
		"cfl biography",
		"skier",
		"model",
		"judge",
		"monarch",
		"speed skater",
		"college football player",
		"economist",
		"medical person",
		"serial killer",
		"gaelic athletic association player",
		"field hockey player",
		"sailor",
		"youtube personality",
		"playboy playmate",
		"nascar driver",
		"volleyball player",
		"motorcycle rider",
		"speedway rider",
		"chef",
		"bishopstyles",
		"horseracing personality",
		"clergy",
		"npb player",
		"darts player",
		"football official",
		"engineer",
		"squash player",
		"scholar",
		"table tennis player",
		"president",
		"president styles",
		"cfl player",
		"jewish leader",
		"governor",
		"mayor",
		"astronaut",
		"poker player",
		"snooker player",
		"indian politician",
		"presenter",
		"actor",
		"dancer",
		"mlb player",
		"fencer",
		"police officer",
		"netball biography",
		"fashion designer",
		"sumo wrestler",
		"nba biography",
		"prime minister",
		//"wrestling team",
		"ice hockey biography",
		"chess biography",
		"pharaoh",
		"ambassador",
		"lacrosse player",
		"hindu leader",
		"le mans driver",
		"musician",
		"latter day saint biography",
		"bishop styles",
		"water polo biography",
		"wrc driver",
		"minister",
		"canadianmp",
		"biathlete",
		"state representative",
		"cardinal styles",
		"bodybuilder",
		"aviator",
		"classical composer",
		"shogi professional",
		"ski jumper",
		"theologian",
		"spy",
		"bishop",
		"go player",
		"climber",
		"soldier",
		"murderer",
		"egyptian dignitary",
		"patriarch",
		"author",
		"champ car driver",
		"pro gaming player",
		"pool player",
		"surfer",
		"equestrian",
		"congressman",
		"sport wrestler",
		"motocross rider",
		"twitch streamer",
		"lds biography",
		"actress",
		"biography",
		"itf women",
		"comedian",
		"video game player",
		"basketball player",
		"triathlete",
		"rebbe",
		"journalist",
		"first lady",
		"archbishop",
		"state senator",
		"mass murderer",
		"bandy biography",
		"roman emperor",
		"sports announcer",
		"historian",
		"mountaineer",
		//

	}
}
