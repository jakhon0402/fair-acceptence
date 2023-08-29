package courseType

type CourseName string

const (
	LANGUAGES       CourseName = "LANGUAGES"
	FOR_ABITURIENTS CourseName = "FOR_ABITURIENTS"
	FOR_KIDS        CourseName = "FOR_KIDS"
	IT              CourseName = "IT"
)

var CoursesNameDescription = map[CourseName]string{
	LANGUAGES:       "Til kursi",
	FOR_ABITURIENTS: "Abituriyentlar uchun",
	FOR_KIDS:        "Bolalar uchun",
	IT:              "IT kursi",
}

var CoursesNames = map[CourseName]string{
	LANGUAGES:       "❇️ Til kurslarimiz",
	FOR_ABITURIENTS: "✳️ Abituriyentlar uchun kurslar",
	FOR_KIDS:        "🌈 Bolalar uchun kurslar",
	IT:              "💻 IT kurslarimiz",
}
