package statStruct

type AuthorStat struct {
	CommitCount          int
	AverageCommitsPerDay float64
	AppendLines          int
	RemoveLines          int
	ModifiedFiles        int
	AddedFiles           int
	DeletedFiles         int
	StartDate            string
	EndDate              string
}

type AuthorTimeStat map[string]AuthorStat      //map[time]AuthorStat
type AuthorTimeStats map[string]AuthorTimeStat //map[user]AuthorTimeStat

type AuthorNameStat struct {
	Author string
	Stat   AuthorStat
}

type YearStats struct {
	Year  int
	Stats []AuthorNameStat
}

type QuarterStats struct {
	Year    int
	Quarter int
	Stats   []AuthorNameStat
}

type MonthStats struct {
	Year  int
	Month int
	Stats []AuthorNameStat
}

type WeekStats struct {
	Year  int
	Week  int
	Stats []AuthorNameStat
}

type CustomStats struct {
	StartDate string
	EndDate   string
	Stats     []AuthorNameStat
}

type ChartData struct {
	XAxis  string
	Series string
}
