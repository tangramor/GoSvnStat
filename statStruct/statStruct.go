package statStruct

type AuthorStat struct {
	CommitCount   int
	AppendLines   int
	RemoveLines   int
	ModifiedFiles int
	AddedFiles    int
	DeletedFiles  int
}

type AuthorTimeStat map[string]AuthorStat      //map[time]AuthorStat
type AuthorTimeStats map[string]AuthorTimeStat //map[user]AuthorTimeStat
type ChartData struct {
	XAxis  string
	Series string
}
