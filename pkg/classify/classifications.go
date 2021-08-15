package classify

type Classification int

const (
	// Unknown represents articles we are unable to classify for some reason.
	Unknown Classification = iota

	// 13692155.
	Philosophy

	// 22921.
	Psychology

	// 25414.
	Religion

	// 10772350.
	History

	// 18963910.
	Geography

	// 26781.
	SocialSciences

	// 24388.
	PoliticalScience

	// 18949668.
	Law

	// 9252.
	Education

	// 18839.
	Music

	// 90317.
	FineArts

	// 23193.
	Philology

	// 22760983.
	Linguistics

	// 18963870.
	Literature

	// 26700.
	Science

	// 18957.
	Medicine

	// 627.
	Agriculture

	// 29816.
	Technology

	// 86586.
	MilitaryScience

	// 149354.
	InformationScience
)
