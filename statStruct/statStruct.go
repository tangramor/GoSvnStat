package statStruct

type AuthorStat struct {
	CommitCount          int
	AverageCommitsPerDay int
	AppendLines          int
	RemoveLines          int
	ModifiedFiles        int
	AddedFiles           int
	DeletedFiles         int
}

type AuthorTimeStat map[string]AuthorStat      //map[time]AuthorStat
type AuthorTimeStats map[string]AuthorTimeStat //map[user]AuthorTimeStat

type AuthorYearStat struct {
	Author string
	Year   int
	Stat   AuthorStat
}

type AuthorQuarterStat struct {
	Author  string
	Year    int
	Quarter int
	Stat    AuthorStat
}

type AuthorMonthStat struct {
	Author string
	Year   int
	Month  int
	Stat   AuthorStat
}

type AuthorWeekStat struct {
	Author string
	Year   int
	Week   int
	Stat   AuthorStat
}

type YearStats struct {
	Year  int
	Stats []AuthorYearStat
}

type QuarterStats struct {
	Year    int
	Quarter int
	Stats   []AuthorQuarterStat
}

type MonthStats struct {
	Year  int
	Month int
	Stats []AuthorMonthStat
}

type WeekStats struct {
	Year  int
	Week  int
	Stats []AuthorWeekStat
}

type ChartData struct {
	XAxis  string
	Series string
}
