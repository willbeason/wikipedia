package classify

func Base() *ClassifiedArticles {
	return &ClassifiedArticles{
		Articles: map[uint32]Classification{
			13692155: Classification_PHILOSOPHY,
			22921:    Classification_PSYCHOLOGY,
			25414:    Classification_RELIGION,
			10772350: Classification_HISTORY,
			18963910: Classification_GEOGRAPHY,
			26781:    Classification_SOCIAL_SCIENCES,
			24388:    Classification_POLITICAL_SCIENCE,
			18949668: Classification_LAW,
			9252:     Classification_EDUCATION,
			18839:    Classification_MUSIC,
			90317:    Classification_FINE_ARTS,
			23193:    Classification_PHILOLOGY_AND_LINGUISTICS, // Philology
			22760983: Classification_PHILOLOGY_AND_LINGUISTICS, // Linguistics
			18963870: Classification_LITERATURE,
			26700:    Classification_SCIENCE,
			22939:    Classification_SCIENCE, // Physics
			25202:    Classification_SCIENCE, // Quantum mechanics
			18957:    Classification_MEDICINE,
			627:      Classification_AGRICULTURE,
			29816:    Classification_TECHNOLOGY,
			86586:    Classification_MILITARY_SCIENCE,
			149354:   Classification_INFORMATION_SCIENCE,
		},
	}
}
